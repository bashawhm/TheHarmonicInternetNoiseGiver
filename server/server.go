package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/pions/webrtc"
	"github.com/pions/webrtc/examples/util"
	"github.com/pions/webrtc/pkg/datachannel"
	"github.com/pions/webrtc/pkg/ice"
)

type Song struct {
	audio  *os.File
	title  string
	artist string
	tag1   string
	tag2   string
}

type Client struct {
	control       net.Conn                  //Defines the control channel
	rtcconn       *webrtc.RTCPeerConnection //Defines the webRTC connection used for managing
	channel       *webrtc.RTCDataChannel    //Defines the webrtc data channel used for file transfer
	username      string
	moderator     bool
	notifications []string
	delay         time.Time
}

type Lobby struct {
	name      string
	admin     Client
	users     []Client
	songQueue []Song
	newUsers  chan Client
}

func (lobby *Lobby) promoteUser(user Client) {
	for i := 0; i < len(lobby.users); i++ {
		if lobby.users[i].username == user.username {
			lobby.users[i].moderator = true
		}
	}
}

func (lobby *Lobby) addSongToQueue(song Song) {
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

func (lobby *Lobby) msgSend(msg string) {
	clients := lobby.getClients()
	for _, client := range clients {
		client.channel.Send(datachannel.PayloadString{Data: []byte(msg)})
	}
}

func (lobby *Lobby) fileSend(file *os.File) {
	clients := lobby.getClients()
	fileStat, _ := file.Stat()
	var fileData []byte = make([]byte, fileStat.Size())
	file.Read(fileData)
	for _, client := range clients {
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

	dataChannel.OnOpen(func() { fmt.Println("Data Channel opened to " + userName) })
	pconn.OnICEConnectionStateChange(func(connState ice.ConnectionState) {
		// fmt.Println(userName + " " + connState.String())
	})

	// fmt.Println("Exchanging offers")
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
		select {
		case newUser := <-lobby.newUsers:
			lobby.users = append(lobby.users, newUser)
		default:
			// fmt.Println(lobby)
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
		fmt.Println("Accepting control connection...")
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
		newLobby := Lobby{name: lobbyName, admin: createClient(config, username, cconn, true), newUsers: newChan}
		lobbies = append(lobbies, newLobby)
		go lobbies[len(lobbies)-1].lobbyHandler()
	}
}
