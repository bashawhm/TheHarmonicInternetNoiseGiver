package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/pions/webrtc"
	"github.com/pions/webrtc/examples/util"
	"github.com/pions/webrtc/pkg/ice"
)

func main() {
	rand.Seed(int64(time.Now().Nanosecond()))
	fmt.Println("THING TEST CLIENT STARTED...")

	webrtc.RegisterDefaultCodecs()
	config := webrtc.RTCConfiguration{
		IceServers: []webrtc.RTCIceServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	pconn, err := webrtc.New(config)
	if err != nil {
		panic(err)
	}
	pconn.OnICEConnectionStateChange(func(connState ice.ConnectionState) {
		fmt.Println(connState.String())
	})

	fmt.Println("Starting control channel...")
	sconn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		panic(err)
	}
	defer sconn.Close()

	if len(os.Args) > 1 {
		fmt.Fprintf(sconn, "TESTLOBBY "+os.Args[1]+"\n")
	} else {
		fmt.Fprintf(sconn, "TESTLOBBY TESTUSER\n")
	}

	nin := bufio.NewScanner(bufio.NewReader(sconn))
	nin.Split(bufio.ScanLines)
	nin.Scan()
	if nin.Text() != "OKAY" {
		panic(nin.Text())
	}
	nin.Scan()
	sd := util.Decode(nin.Text())
	fmt.Println("DECODED")

	offer := webrtc.RTCSessionDescription{
		Type: webrtc.RTCSdpTypeOffer,
		Sdp:  string(sd),
	}
	err = pconn.SetRemoteDescription(offer)
	if err != nil {
		panic(err)
	}

	answer, err := pconn.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(sconn, util.Encode(answer.Sdp)+"\n")

	/*
		f, err := os.Create("/tmp/song-" + strconv.Itoa(rand.Int()) + ".wav")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		pconn.OnDataChannel(func(d *webrtc.RTCDataChannel) {
			d.OnOpen(func() {
				fmt.Println("Opened data connection to server")
			})

			d.OnMessage(func(payload datachannel.Payload) {
				switch p := payload.(type) {
				case *datachannel.PayloadString:
					// fmt.Printf("Message '%s' from DataChannel '%s' payload '%s'\n", p.PayloadType().String(), d.Label, string(p.Data))
					f.Write(p.Data)
				case *datachannel.PayloadBinary:
					// fmt.Printf("Message '%s' from DataChannel '%s' payload '% 02x'\n", p.PayloadType().String(), d.Label, p.Data)
					f.Write(p.Data)
				default:
					fmt.Printf("Message '%s' from DataChannel '%s' no payload \n", p.PayloadType().String(), d.Label)
				}
			})
		})
	*/
	select {}
}
