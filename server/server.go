package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/pions/webrtc"
	"github.com/pions/webrtc/examples/util"
	"github.com/pions/webrtc/pkg/datachannel"
	"github.com/pions/webrtc/pkg/ice"
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

func debugPrintln(debugLevel int, a ...interface{}) {
	if DEBUGLEVEL >= debugLevel {
		fmt.Print("[" + time.Now().Format("15:04:05") + "] ")
		fmt.Println(a...)
	}
}

type Song struct {
	audio  *os.File
	title  string
	artist string
	tag1   string
	tag2   string
}

type Client struct {
	control       net.Conn                  //Defines the control channel
	rtcconn       *webrtc.RTCPeerConnection //Defines the webRTC connection used for managing data channel
	channel       *webrtc.RTCDataChannel    //Defines the webrtc data channel used for file transfer
	username      string
	moderator     bool
	notifications []string
	delay         time.Time
}

type Lobby struct {
	name          string
	admin         Client
	users         []Client
	songQueue     []Song
	newUsers      chan Client
	bufferedUsers []Client
	userAccept    chan Client
}

func (lobby *Lobby) promoteUser(user Client) {
	for i := 0; i < len(lobby.users); i++ {
		if lobby.users[i].username == user.username {
			debugPrintln(Dump, "Promoting: "+lobby.users[i].username+" to moderator")
			lobby.users[i].moderator = true
		}
	}
}

func (lobby *Lobby) addSongToQueue(song Song) {
	debugPrintln(Dump, "adding "+song.title+" to song queue")
	lobby.songQueue = append(lobby.songQueue, song)
}

func (lobby *Lobby) getClients() []Client {
	var clients []Client
	clients = append(clients, lobby.admin)
	for _, client := range lobby.users {
		clients = append(clients, client)
	}
	return clients
}

/*
func (client *Client) getDelay() time.Time {

}

func (lobby *Lobby) pushSong(song Song) string {

}

func (lobby *Lobby) syncPlay(song Song) string {

}

func (lobby *Lobby) syncPause() string {

}
*/

func (lobby *Lobby) sendNotifications() {

}

func (lobby *Lobby) msgSend(msg string) {
	clients := lobby.getClients()
	for _, client := range clients {
		debugPrintln(Dump, "Sending "+msg+" to client "+client.username)
		client.channel.Send(datachannel.PayloadString{Data: []byte(msg)})
	}
}

func (lobby *Lobby) fileSend(file *os.File) {
	clients := lobby.getClients()
	fileStat, _ := file.Stat()
	var fileData []byte = make([]byte, fileStat.Size())
	file.Read(fileData)
	for _, client := range clients {
		debugPrintln(Dump, "Sending file to client "+client.username)
		//Tell glient begining of file
		fmt.Fprintf(client.control, "SEND\n")
		for i := 0; i < len(fileData); i += 1000 {
			if (i + 1000) > len(fileData) {
				err := client.channel.Send(datachannel.PayloadBinary{Data: fileData[i:]})
				if err != nil {
					panic(err)
				}
			} else {
				err := client.channel.Send(datachannel.PayloadBinary{Data: fileData[i : i+1000]})
				time.Sleep(time.Microsecond * 50) //May need to be higher on slower networks
				if err != nil {
					panic(err)
				}
			}
		}
		//Tell client end of file
		fmt.Fprintf(client.control, "OKAY\n")
	}
}

func createClient(config webrtc.RTCConfiguration, userName string, conn net.Conn, mod bool) Client {
	if mod {
		debugPrintln(Spew, "Creating client "+userName+" with mod status true")
	} else {
		debugPrintln(Spew, "Creating client "+userName+" with mod status false")
	}
	//Handle webRTC stuff
	pconn, err := webrtc.New(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create new connection\n")
		panic(err)
	}

	dataChannel, err := pconn.CreateDataChannel("audio", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create data channel\n")
		panic(err)
	}

	dataChannel.OnOpen(func() { debugPrintln(Dump, "Data Channel opened to "+userName) })
	pconn.OnICEConnectionStateChange(func(connState ice.ConnectionState) {
		debugPrintln(Dump, userName+" "+connState.String())
	})

	debugPrintln(Spew, "Exchanging offers")
	offer, err := pconn.CreateOffer(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create offer\n")
		panic(err)
	}
	fmt.Fprintf(conn, util.Encode(offer.Sdp)+"\n")
	nin := bufio.NewScanner(bufio.NewReader(conn))
	nin.Split(bufio.ScanLines)
	nin.Scan()
	sd := util.Decode(nin.Text())

	answer := webrtc.RTCSessionDescription{
		Type: webrtc.RTCSdpTypeAnswer,
		Sdp:  sd,
	}
	err = pconn.SetRemoteDescription(answer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set remote descriptor\n")
		panic(err)
	}

	return Client{username: userName, control: conn, rtcconn: pconn, channel: dataChannel, moderator: mod}
}

func (lobby *Lobby) lobbyHandler() {
	for {
		lobby.sendNotifications()
		select {
		case newUser := <-lobby.newUsers:
			lobby.bufferedUsers = append(lobby.bufferedUsers, newUser)
			clients := lobby.getClients()
			debugPrintln(Dump, "Sending join notifications")
			for i := 0; i < len(clients); i++ {
				if clients[i].moderator {
					debugPrintln(Spew, "sending join notification "+newUser.username+" to "+clients[i].username)
					clients[i].notifications = append(clients[i].notifications, "User "+newUser.username+" wants to join "+lobby.name)
				}
			}
		case accept := <-lobby.userAccept:
			debugPrintln(Dump, accept.username+" Joinging "+lobby.name)
			lobby.users = append(lobby.users, accept)
		default:
			// fmt.Println(lobby)
			//Handle all client commands
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
					debugPrintln(Spew, err)
					continue
				}
				if n == 0 {
					continue
				}
				//Something was read, handle connection in a non blocking way
				if lobby.admin.username == clients[i].username { //If send from admin
					//If accepting users move them from bufferedUsers to userAccept channel

				} else if clients[i].moderator { //If sent from a moderator
					//If accepting users move them from bufferedUsers to userAccept channel

				} else { //Sent by user

				}
			}
		}
	}
}

func main() {
	fmt.Fprintf(os.Stderr, "THING server starting...\n")
	webrtc.RegisterDefaultCodecs()
	config := webrtc.RTCConfiguration{
		IceServers: []webrtc.RTCIceServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	var lobbies []Lobby
	ln, err := net.Listen("tcp", "localhost:9000")
	if err != nil {
		panic(err)
	}
	for {
	start:
		debugPrintln(Info, "Accepting control connection...")
		cconn, err := ln.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to accept connection")
			cconn.Close()
			continue
		}
		nin := bufio.NewScanner(bufio.NewReader(cconn))
		nin.Split(bufio.ScanWords)
		//Format:
		//LOBBYNAME USERNAME
		nin.Scan()
		lobbyName := nin.Text()
		nin.Scan()
		username := nin.Text()
		//Check if username is taken
		for i := 0; i < len(lobbies); i++ {
			clients := lobbies[i].getClients()
			for j := 0; j < len(clients); j++ {
				if clients[j].username == username {
					fmt.Fprintf(cconn, username+" TAKEN\n")
					cconn.Close()
					goto start
				}
			}
		}
		fmt.Fprintf(cconn, "OKAY\n")

		for i := 0; i < len(lobbies); i++ {
			if lobbies[i].name == lobbyName {
				select {
				case lobbies[i].newUsers <- createClient(config, username, cconn, false):
				default:
					fmt.Println("Channel Full, rejecting user")
				}
				goto start //Sorry, but it really is the simplest solution
			}
		}
		//If the lobby doesn't exist, create it and spawn handler
		newChan := make(chan Client, 25)
		acceptChan := make(chan Client, 25)
		newLobby := Lobby{name: lobbyName, admin: createClient(config, username, cconn, true), newUsers: newChan, userAccept: acceptChan}
		lobbies = append(lobbies, newLobby)
		go lobbies[len(lobbies)-1].lobbyHandler()
	}
}
