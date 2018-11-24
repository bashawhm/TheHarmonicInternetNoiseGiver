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

func (lobby *Lobby) massSend(msg string) {
	clients := lobby.getClients()
	for _, client := range clients {
		client.channel.Send(datachannel.PayloadString{Data: []byte(msg)})
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

	var lobby Lobby = Lobby{name: "Testing lobby"}

	//Testing function
	go func() {
		cin := bufio.NewScanner(os.Stdin)
		cin.Split(bufio.ScanLines)
		for {
			cin.Scan()
			lobby.massSend(cin.Text())
		}
	}()

	ln, err := net.Listen("tcp", "localhost:9000")
	var admin bool = true
	for {
		fmt.Println("Accepting control connections...")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get connection")
			continue
		}
		cconn, err := ln.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to accept connection")
			continue
		}

		//Create a new WebRTC peer Connection
		pconn, err := webrtc.New(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create new connection\n")
			continue
		}
		defer pconn.Close()
		maxRetrans := uint16(8)
		dataChannel, err := pconn.CreateDataChannel("audio", &webrtc.RTCDataChannelInit{MaxRetransmits: &maxRetrans})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create data channel\n")
			continue
		}

		dataChannel.OnOpen(func() { fmt.Println("Data Channel opened") })

		pconn.OnICEConnectionStateChange(func(connState ice.ConnectionState) {
			fmt.Println(connState.String())
		})

		fmt.Println("Exchanging offers")
		offer, err := pconn.CreateOffer(nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create offer\n")
			continue
		}
		fmt.Fprintf(cconn, util.Encode(offer.Sdp)+"\n")
		nin := bufio.NewScanner(bufio.NewReader(cconn))
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
			continue
		}

		if admin {
			lobby.admin = Client{control: cconn, rtcconn: pconn, channel: dataChannel}
			admin = false
		} else {
			lobby.users = append(lobby.users, Client{control: cconn, rtcconn: pconn, channel: dataChannel})
		}
	}
}
