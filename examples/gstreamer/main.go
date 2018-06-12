package main

import (
	"fmt"
	"os"

	"bufio"
	"encoding/base64"
	"github.com/pions/webrtc"
	"github.com/pions/webrtc/examples/gstreamer/gst"
	"github.com/pions/webrtc/pkg/rtp"
	"sync/atomic"
)

var trackCount uint64

func startWebrtc(pipeline *gst.Pipeline) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Browser base64 Session Description: ")
	rawSd, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	fmt.Println("\nGolang base64 Session Description: ")

	sd, err := base64.StdEncoding.DecodeString(rawSd)
	if err != nil {
		panic(err)
	}

	/* Everything below is the pion-WebRTC API, thanks for using it! */

	// Create a new RTCPeerConnection
	peerConnection := &webrtc.RTCPeerConnection{}

	// Set a handler for when a new remote track starts, this handler starts a gstreamer pipeline
	// with the first track and assumes it is VP8 video data.
	peerConnection.Ontrack = func(mediaType webrtc.MediaType, packets chan *rtp.Packet) {
		go func() {
			track := atomic.AddUint64(&trackCount, 1)
			fmt.Printf("Track %d has started \n", track)
			if track == 1 && mediaType == webrtc.VP8 {
				for {
					p := <-packets
					pipeline.Push(p.Raw)
				}
			}
		}()
	}

	// Set the remote SessionDescription
	if err := peerConnection.SetRemoteDescription(string(sd)); err != nil {
		panic(err)
	}

	// Sets the LocalDescription, and starts our UDP listeners
	if err := peerConnection.CreateOffer(); err != nil {
		panic(err)
	}

	// Get the LocalDescription and take it to base64 so we can paste in browser
	localDescriptionStr := peerConnection.LocalDescription.Marshal()
	fmt.Println(base64.StdEncoding.EncodeToString([]byte(localDescriptionStr)))
	select {}
}

func main() {
	p := gst.CreatePipeline()
	go startWebrtc(p)
	p.Start()
}