package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/pions/webrtc"
	"github.com/pions/webrtc/examples/util"
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
	control       net.Conn
	rtcconn       *webrtc.RTCPeerConnection
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

	var clients []Client
	ln, err := net.Listen("tcp", "localhost:9000")
	for {
		fmt.Println("Accepting control connections...")
		if err != nil {
			fmt.Println("Failed to get connection")
			continue
		}
		cconn, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		//Create a new WebRTC peer Connection
		pconn, err := webrtc.New(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create new connection\n")
		}
		defer pconn.Close()

		pconn.OnICEConnectionStateChange(func(connState ice.ConnectionState) {
			fmt.Println(connState.String())
		})

		fmt.Fprintf(os.Stderr, "Creating offer\n")
		offer, err := pconn.CreateOffer(nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create offer\n")
		}
		fmt.Fprintf(cconn, util.Encode(offer.Sdp)+"\n")

		nin := bufio.NewScanner(bufio.NewReader(cconn))
		nin.Split(bufio.ScanLines)
		nin.Scan()
		fmt.Println("Got offer from client")
		sd := util.Decode(nin.Text())

		answer := webrtc.RTCSessionDescription{
			Type: webrtc.RTCSdpTypeAnswer,
			Sdp:  sd,
		}

		err = pconn.SetRemoteDescription(answer)
		if err != nil {
			panic(err)
		}

		clients = append(clients, Client{control: cconn, rtcconn: pconn})

	}
}
