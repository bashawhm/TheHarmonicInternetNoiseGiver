package main

import (
	"bufio"
	"fmt"
	"net"

	"github.com/pions/webrtc"
	"github.com/pions/webrtc/examples/util"
	"github.com/pions/webrtc/pkg/ice"
)

func main() {
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

	sconn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		panic(err)
	}
	defer sconn.Close()
	nin := bufio.NewScanner(bufio.NewReader(sconn))
	nin.Split(bufio.ScanLines)
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
	select {}
}
