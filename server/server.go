package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pions/webrtc"
	"github.com/pions/webrtc/pkg/datachannel"
	"github.com/pions/webrtc/pkg/ice"
	"golang.org/x/net/websocket"
)

//Debug levels
const (
	None = iota
	Info
	Dump
	Spew
)

//DEBUGLEVEL represents the amount of debug output that will be generated by the server
const DEBUGLEVEL = Spew

//Global state of all the lobbies in the server
var lobbies []Lobby

//A mutex to prevent threading mayhem on lobbies
var lobbyMutex = &sync.Mutex{}

func debugPrintln(debugLevel int, a ...interface{}) {
	if DEBUGLEVEL >= debugLevel {
		fmt.Print("[" + time.Now().Format("15:04:05") + "] ")
		fmt.Println(a...)
	}
}

//Song is the server-side representation of a song used to manage the songs uploaded to the lobby
type Song struct {
	audio  []byte
	title  string
	artist string
	tag1   string
	tag2   string
}

//Client is the server-side view of a THING client used to manage connections and permissions
type Client struct {
	control       *websocket.Conn           //Defines the control channel
	rtcconn       *webrtc.RTCPeerConnection //Defines the webRTC connection used for managing data channel
	channel       *webrtc.RTCDataChannel    //Defines the webrtc data channel used for file transfer
	username      string
	moderator     bool
	notifications []string
	delay         time.Duration
}

//Lobby is the server-side representation of a collection of clients and songs
type Lobby struct {
	name          string
	admin         Client
	users         []Client
	songQueue     []Song
	newUsers      chan Client //Clients yet to be buffered
	bufferedUsers []Client    //Clients buffered until accepted
	userAccept    chan Client //Clients yet to be accepted
	partialSong   []byte      //If there is a song sending
	partialMut    *sync.Mutex
}

type THING struct {
	Command string `json:"command"`
}

//Promotes a user to moderator
func (lobby *Lobby) promoteUser(user Client) {
	for i := 0; i < len(lobby.users); i++ {
		if lobby.users[i].username == user.username {
			debugPrintln(Dump, "Promoting: "+lobby.users[i].username+" to moderator")
			lobby.users[i].moderator = true
		}
	}
}

//Adds song to the lobbies queue
func (lobby *Lobby) addSongToQueue(song Song) {
	debugPrintln(Dump, "adding "+song.title+" to song queue")
	lobby.songQueue = append(lobby.songQueue, song)
}

//Returns all the clients in a particular lobby, excluding the buffered clients
func (lobby *Lobby) getClients() []Client {
	var clients []Client
	clients = append(clients, lobby.admin)
	for _, client := range lobby.users {
		clients = append(clients, client)
	}
	return clients
}

func (client *Client) updateDelayTime() {
	t1 := time.Now()
	var packet THING
	packet.Command = "PING\n"
	websocket.JSON.Send(client.control, packet)
	websocket.JSON.Receive(client.control, &packet)
	t2 := time.Now()
	diff := t2.Sub(t1)
	client.delay = diff
}

func (lobby *Lobby) pushSong(song Song) {
	lobby.fileSend(song)
}

func (lobby *Lobby) syncPlay(song Song) {
	clients := lobby.getClients()
	for i := 0; i < len(clients); i++ {
		clients[i].updateDelayTime()
	}
	max := clients[0].delay
	for i := 1; i < len(clients); i++ {
		if clients[i].delay > max {
			max = clients[i].delay
		}
	}
	var packet THING
	packet.Command = "PLAY " + song.title + "\n"
	for j := 0; j < len(clients); j++ {
		go func(i int) {
			time.Sleep(max - clients[i].delay)
			websocket.JSON.Send(clients[i].control, packet)
		}(j)
	}
}

func (lobby *Lobby) syncPause() {
	clients := lobby.getClients()
	for i := 0; i < len(clients); i++ {
		clients[i].updateDelayTime()
	}
	max := clients[0].delay
	for i := 1; i < len(clients); i++ {
		if clients[i].delay > max {
			max = clients[i].delay
		}
	}
	var packet THING
	packet.Command = "PAUSE\n"
	for j := 0; j < len(clients); j++ {
		go func(i int) {
			time.Sleep(max - clients[i].delay)
			websocket.JSON.Send(clients[i].control, packet)
		}(j)
	}
}

func (lobby *Lobby) sendNotifications() {
	clients := lobby.getClients()
	var packet THING
	for i := 0; i < len(clients); i++ {
		packet.Command = "NOTIFY "
		for j := 0; j < len(clients[i].notifications); i++ {
			packet.Command += clients[i].notifications[j] + "\n"
		}
		err := websocket.JSON.Send(clients[i].control, packet)
		if err != nil {
			debugPrintln(Info, err)
		}
	}
}

func (lobby *Lobby) updateSong(song Song) {
	clients := lobby.getClients()
	var packet THING
	packet.Command = "UPDATE " + song.title + " " + song.artist + " " + song.tag1 + " " + song.tag2 + "\n"
	for i := 0; i < len(clients); i++ {
		websocket.JSON.Send(clients[i].control, packet)
	}
}

//Used to send song data to all connected clients over a WebRTC data channel
func (lobby *Lobby) fileSend(song Song) {
	clients := lobby.getClients()
	var packet THING
	for _, client := range clients {
		debugPrintln(Dump, "Sending file to client "+client.username)
		//Tell client a transfer is starting
		packet.Command = "SEND " + song.title + " " + song.artist + " " + song.tag1 + " " + song.tag2 + "\n"
		websocket.JSON.Send(client.control, packet)
		//Transfer in >=1000 byte chunks because of the limits of WebRTC
		for i := 0; i < len(song.audio); i += 1000 {
			if (i + 1000) > len(song.audio) { //The last set of bytes in the file
				err := client.channel.Send(datachannel.PayloadBinary{Data: song.audio[i:]})
				if err != nil {
					panic(err)
				}
			} else {
				err := client.channel.Send(datachannel.PayloadBinary{Data: song.audio[i : i+1000]})
				time.Sleep(time.Microsecond * 50) //May need to be higher on slower networks
				if err != nil {
					panic(err)
				}
			}
		}
		//Tell client end of transmision
		packet.Command = "OKAY\n"
		websocket.JSON.Send(client.control, packet)
	}
}

func (lobby *Lobby) fileRecv(p datachannel.Payload) {
	lobby.partialMut.Lock()
	var audio []byte
	switch payload := p.(type) {
	case *datachannel.PayloadBinary:
		audio = append(audio, payload.Data...)
	case *datachannel.PayloadString:
		audio = append(audio, payload.Data...)
	default:
		debugPrintln(Info, "Failed to recieve song data, not binary or string")
		debugPrintln(Dump, "Payload type "+p.PayloadType().String())
		return
	}
	lobby.partialSong = append(lobby.partialSong, audio...)
	lobby.partialMut.Unlock()
}

//Creates a Client and handles WebRTC magic
func createClient(config webrtc.RTCConfiguration, userName string, conn *websocket.Conn, mod bool) Client {
	//Debug printing...
	if mod {
		debugPrintln(Spew, "Creating client "+userName+" with mod status true")
	} else {
		debugPrintln(Spew, "Creating client "+userName+" with mod status false")
	}

	//Creates the webRTC peer connection with a config to point it at google's public STUN server
	pconn, err := webrtc.New(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create new connection\n")
		panic(err)
	}
	//Create the data channel from the peer connection
	//The data channel named "audio" is where the file data will actually be sent
	//This is distinct from the peer connection
	dataChannel, err := pconn.CreateDataChannel("audio", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create data channel\n")
		panic(err)
	}

	//When the data channel opens up print to debug command line
	dataChannel.OnOpen(func() { debugPrintln(Dump, "Data Channel opened to "+userName) })
	//When the peer connection is checking or fully connects it'll call this handler and print to the debug log
	//Generally unimportant, could live without
	pconn.OnICEConnectionStateChange(func(connState ice.ConnectionState) {
		debugPrintln(Dump, userName+" "+connState.String())
	})

	//Create WebRTC offer to link up with the client
	debugPrintln(Spew, "Exchanging offers")
	offer, err := pconn.CreateOffer(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create offer\n")
		panic(err)
	}
	var packet THING
	packet.Command = offer.Sdp
	//Send offer to client
	websocket.JSON.Send(conn, packet)
	//Get the clients offer back to link the server with the client
	websocket.JSON.Receive(conn, &packet)
	sd := packet.Command

	//Create answer out of the clients offer to fully create the connection
	answer := webrtc.RTCSessionDescription{
		Type: webrtc.RTCSdpTypeAnswer,
		Sdp:  sd,
	}
	//Actually fully connects the client and the server
	err = pconn.SetRemoteDescription(answer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set remote descriptor\n")
		panic(err)
	}
	//Construct client with everything
	return Client{username: userName, control: conn, rtcconn: pconn, channel: dataChannel, moderator: mod}
}

//Handles all operations within the lobby
func (lobby *Lobby) lobbyHandler() {
	for {
		select {
		//If a client wants to connect
		case newUser := <-lobby.newUsers:
			lobby.bufferedUsers = append(lobby.bufferedUsers, newUser)
			clients := lobby.getClients()
			debugPrintln(Dump, "Sending join notifications")
			//Send join notifications to the admin client and all moderator clients
			for i := 0; i < len(clients); i++ {
				if clients[i].moderator { //Admin is also a moderator
					debugPrintln(Spew, "sending join notification "+newUser.username+" to "+clients[i].username)
					clients[i].notifications = append(clients[i].notifications, "User "+newUser.username+" wants to join "+lobby.name)
				}
			}
		//If a client is pending acception
		case accept := <-lobby.userAccept:
			debugPrintln(Dump, accept.username+" Joinging "+lobby.name)
			lobby.users = append(lobby.users, accept)
		//Handle all client Commands
		default:
			lobby.sendNotifications()
			clients := lobby.getClients()
			for i := 0; i < len(clients); i++ {
				buffer := make([]byte, 8192)
				n, err := clients[i].control.Read(buffer)
				if err == io.EOF {
					debugPrintln(Info, clients[i].username+" disconnected")
					if clients[i].username == lobby.admin.username { //Destroy lobby and disconnect everyone
						debugPrintln(Info, "Admin disconnected, destroying lobby "+lobby.name)
						for j := 0; j < len(clients); j++ {
							clients[j].control.Close()
							clients[i].rtcconn.Close()
						}
						for j := 0; j < len(lobby.bufferedUsers); j++ {
							lobby.bufferedUsers[j].control.Close()
							lobby.bufferedUsers[j].rtcconn.Close()
						}
						//Ideally deallocate lobby, TODO: find better way to invalidate it
						lobby.name = ""
						lobby.admin = Client{}
						lobby.users = nil
						lobby.bufferedUsers = nil
						return
					}
				}
				if err != nil {
					// debugPrintln(Spew, err)
					continue
				}
				if n == 0 {
					continue
				}
				var input THING
				err = json.Unmarshal(buffer, &input)
				if err != nil {
					debugPrintln(Info, err)
					continue
				}
				//If we get here there are no network errors
				sin := bufio.NewScanner(strings.NewReader(input.Command))
				sin.Split(bufio.ScanWords)
				var packet THING
				//Something was read, handle connection in a non blocking way
				if lobby.admin.username == clients[i].username { //If send from admin
					//If accepting users move them from bufferedUsers to userAccept channel
					res := sin.Scan()
					if !res {
						debugPrintln(Dump, sin.Err())
						continue
					}
					switch sin.Text() {
					case "ACCEPT":
						res := sin.Scan()
						if !res {
							debugPrintln(Dump, sin.Err())
							continue
						}
						user := sin.Text()
						for j := 0; j < len(lobby.bufferedUsers); i++ {
							if lobby.bufferedUsers[j].username == user {
								lobby.userAccept <- lobby.bufferedUsers[j]
								//Remove Client from buffer
								if len(lobby.bufferedUsers) == 1 {
									lobby.bufferedUsers = []Client{}
								} else {
									lobby.bufferedUsers = append(lobby.bufferedUsers[:j], lobby.bufferedUsers[j+1:]...)
								}
								packet.Command = "OKAY\n"
								websocket.JSON.Send(clients[i].control, packet)
								break
							}
						}
					case "PLAY":
						packet.Command = "OKAY\n"
						websocket.JSON.Send(clients[i].control, packet)
						res := sin.Scan()
						if !res {
							debugPrintln(Dump, sin.Err())
							continue
						}
						songTitle := sin.Text()
						for _, s := range lobby.songQueue {
							if s.title == songTitle {
								lobby.syncPlay(s)
								break
							}
						}
					case "PAUSE":
						lobby.syncPause()
					case "SONG":
						res := sin.Scan()
						if !res {
							debugPrintln(Dump, sin.Err())
							continue
						}
						songTitle := sin.Text()
						res = sin.Scan()
						if !res {
							debugPrintln(Dump, sin.Err())
							continue
						}
						if sin.Text() != "SET" {
							continue
						}
						res = sin.Scan()
						if !res {
							debugPrintln(Dump, sin.Err())
							continue
						}
						switch sin.Text() {
						case "ARTIST":
							res := sin.Scan()
							if !res {
								debugPrintln(Dump, sin.Err())
								continue
							}
							newArtist := sin.Text()
							for j := 0; j < len(lobby.songQueue); j++ {
								if lobby.songQueue[j].title == songTitle {
									lobby.songQueue[j].artist = newArtist
								}
							}
						case "TAG1":
							res := sin.Scan()
							if !res {
								debugPrintln(Dump, sin.Err())
								continue
							}
							newTag1 := sin.Text()
							for j := 0; j < len(lobby.songQueue); j++ {
								if lobby.songQueue[j].title == songTitle {
									lobby.songQueue[j].tag1 = newTag1
								}
							}
						case "TAG2":
							res := sin.Scan()
							if !res {
								debugPrintln(Dump, sin.Err())
								continue
							}
							newTag2 := sin.Text()
							for j := 0; j < len(lobby.songQueue); j++ {
								if lobby.songQueue[j].title == songTitle {
									lobby.songQueue[j].tag2 = newTag2
								}
							}
						default:
							continue
						}
						for j := 0; j < len(lobby.songQueue); j++ {
							if lobby.songQueue[j].title == songTitle {
								lobby.updateSong(lobby.songQueue[j])
							}
						}
					case "SEND":
						lobby.partialSong = []byte{}
						res := sin.Scan()
						if !res {
							debugPrintln(Dump, "Invalid protocol: no song title on send")
							continue
						}
						songTitle := sin.Text()
						var songArtist string
						var songTag1 string
						var songTag2 string
						res = sin.Scan()
						if res {
							songArtist = sin.Text()
						}
						res = sin.Scan()
						if res {
							songTag1 = sin.Text()
						}
						res = sin.Scan()
						if res {
							songTag2 = sin.Text()
						}
						//Wait for transfer to progress
						var packet THING
						websocket.JSON.Receive(clients[i].control, &packet)
						//When transfer ends
						if packet.Command != "OKAY\n" {
							debugPrintln(Info, "Failed to get OKAY for upload: "+packet.Command)
							continue
						}
						//Fill in other fields
						var song Song
						song.title = songTitle
						song.artist = songArtist
						song.tag1 = songTag1
						song.tag2 = songTag2
						song.audio = lobby.partialSong
						lobby.addSongToQueue(song)
						//Send song to clients
						lobby.pushSong(song)
					default:
					}
				} else if clients[i].moderator { //If sent from a moderator
					//If accepting users move them from bufferedUsers to userAccept channel
					res := sin.Scan()
					if !res {
						debugPrintln(Dump, sin.Err())
						continue
					}
					if sin.Text() == "ACCEPT" {
						res := sin.Scan()
						if !res {
							debugPrintln(Dump, sin.Err())
							continue
						}
						user := sin.Text()
						for j := 0; j < len(lobby.bufferedUsers); i++ {
							if lobby.bufferedUsers[j].username == user {
								lobby.userAccept <- lobby.bufferedUsers[j]
								//Remove Client from buffer
								lobby.bufferedUsers = append(lobby.bufferedUsers[:j], lobby.bufferedUsers[j+1:]...)
								packet.Command = "OKAY\n"
								websocket.JSON.Send(clients[i].control, packet)
								break
							}
						}
					}
				} else { //Sent by user

				}
			}
		}
	}
}

//Web socket handler called by the HTTP server
func THINGServer(cconn *websocket.Conn) {
	fmt.Fprintf(os.Stderr, "THING client connecting...\n")
	// lobbyMutex.Lock()
	config := webrtc.RTCConfiguration{
		IceServers: []webrtc.RTCIceServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	for {
		debugPrintln(Info, "Accepting control connection...")
		var input THING
		err := websocket.JSON.Receive(cconn, &input)
		if err != nil {
			debugPrintln(Dump, err)
			return
		}
		// fmt.Println(input)
		nin := bufio.NewScanner(strings.NewReader(input.Command))
		nin.Split(bufio.ScanWords)
		var packet THING
		for nin.Scan() {
			switch nin.Text() {
			case "LOBBY":
				packet.Command = ""
				for i := 0; i < len(lobbies); i++ {
					packet.Command = packet.Command + lobbies[i].name + "\n"
				}
				packet.Command = "SERVERLIST\n" + packet.Command
				fmt.Println("Sending message:\n" + packet.Command)
				websocket.JSON.Send(cconn, packet)
			case "JOIN":
				res := nin.Scan()
				if !res {
					debugPrintln(Dump, nin.Err())
					continue
				}
				lobbyName := nin.Text()
				res = nin.Scan()
				if !res {
					debugPrintln(Dump, nin.Err())
					continue
				}
				username := nin.Text()
				//Check if username is taken
				for i := 0; i < len(lobbies); i++ {
					clients := lobbies[i].getClients()
					for j := 0; j < len(clients); j++ {
						if clients[j].username == username {
							packet.Command = username + " TAKEN\n"
							websocket.JSON.Send(cconn, packet)
							continue
						}
					}
				}
				//Check if lobby exists and add user to it
				for i := 0; i < len(lobbies); i++ {
					if lobbies[i].name == lobbyName {
						packet.Command = "OKAY\n"
						websocket.JSON.Send(cconn, packet)
						select {
						case lobbies[i].newUsers <- createClient(config, username, cconn, false):
						default:
							fmt.Println("Channel Full, rejecting user")
						}
						packet.Command = "OKAY\n"
						websocket.JSON.Send(cconn, packet)
						// lobbyMutex.Unlock()
						select {}
					}
				}
			case "CREATE":
				//If the lobby doesn't exist, create it and spawn handler
				res := nin.Scan()
				if !res {
					debugPrintln(Dump, nin.Err())
					return
				}
				lobbyName := nin.Text()
				res = nin.Scan()
				if !res {
					debugPrintln(Dump, nin.Err())
					return
				}
				username := nin.Text()
				for i := 0; i < len(lobbies); i++ {
					if lobbies[i].name == lobbyName {
						packet.Command = lobbyName + " TAKEN\n"
						websocket.JSON.Send(cconn, packet)
						return
					}
				}
				packet.Command = "OKAY\n"
				websocket.JSON.Send(cconn, packet)
				newChan := make(chan Client, 25)
				acceptChan := make(chan Client, 25)
				newLobby := Lobby{name: lobbyName, admin: createClient(config, username, cconn, true), newUsers: newChan, userAccept: acceptChan}
				lobbies = append(lobbies, newLobby)
				fmt.Println(lobbies)
				packet.Command = "OKAY\n"
				websocket.JSON.Send(cconn, packet)
				lobbies[len(lobbies)-1].partialMut = new(sync.Mutex)
				lobbies[len(lobbies)-1].admin.channel.OnMessage(lobbies[len(lobbies)-1].fileRecv)
				go lobbies[len(lobbies)-1].lobbyHandler()
				// lobbyMutex.Unlock()
				select {}
			default:
			}
		}
	}
}

func main() {
	//Start HTTP server and send client to browser
	http.Handle("/", http.FileServer(http.Dir("../client/")))
	//Start the web socket handler
	http.Handle("/socket", websocket.Handler(THINGServer))
	debugPrintln(Info, "Starting HTTP and WS server")

	//Initialization stuff
	webrtc.RegisterDefaultCodecs()

	lobbies = append(lobbies, Lobby{name: "SampleLobby"})
	lobbies = append(lobbies, Lobby{name: "SamsLobby"})
	lobbies = append(lobbies, Lobby{name: "SomeParty"})
	lobbies = append(lobbies, Lobby{name: "SomeEvent"})
	lobbies = append(lobbies, Lobby{name: "AnythingElse"})
	lobbies = append(lobbies, Lobby{name: "SampleLobby2"})
	lobbies = append(lobbies, Lobby{name: "SamsLobby2"})
	lobbies = append(lobbies, Lobby{name: "SomeParty2"})
	lobbies = append(lobbies, Lobby{name: "SomeEvent2"})
	lobbies = append(lobbies, Lobby{name: "AnythingElse2"})
	//Start server
	http.ListenAndServe(":80", nil)
}
