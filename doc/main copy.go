package main

import (
	"flag"
	"fmt"

	ilog "github.com/pion/ion-log"
	sdk "github.com/pion/ion-sdk-go"
	"github.com/pion/webrtc/v3"

	// Note: If you don't have a camera or microphone or your adapters are not supported,
	//       you can always swap your adapters with our dummy adapters below.
	// _ "github.com/pion/mediadevices/pkg/driver/videotest"
	// _ "github.com/pion/mediadevices/pkg/driver/audiotest"
	_ "github.com/pion/mediadevices/pkg/driver/camera"     // This is required to register camera adapter
	_ "github.com/pion/mediadevices/pkg/driver/microphone" // This is required to register microphone adapter
)

var (
	log = ilog.NewLoggerWithFields(ilog.DebugLevel, "main", nil)
)

func main() {

	// parse flag
	var session, addr string
	flag.StringVar(&addr, "addr", "192.168.1.199:5551", "ion-sfu grpc addr")
	flag.StringVar(&session, "session", "ion", "join session name")
	flag.Parse()

	connector := sdk.NewConnector(addr)
	rtc, err := sdk.NewRTC(connector)
	if err != nil {
		panic(err)
	}

	rtc.GetPubTransport().GetPeerConnection().OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log.Infof("Connection state changed: %s", state)
	})

	var infos []*sdk.Subscription
	infos = append(infos, &sdk.Subscription{
		TrackId: "ion",
		Layer:   "PiVideoSource",
	})
	// infos[0] = &sdk.Subscription{
	// 	TrackId: "ion",
	// }
	err = rtc.Subscribe(infos)
	//err = rtc.Join(session, "videoSource"+sdk.RandomKey(8))

	if err != nil {
		log.Errorf("join err=%v", err)
		panic(err)
	}
	rtc.OnDataChannel = func(dc *webrtc.DataChannel) {
		log.Infof("c.OnDataChannel = func")

		vpxParams, err := vpx.NewVP8Params()
		if err != nil {
			panic(err)
		}
		vpxParams.BitRate = 500_000 // 500kbps

		codecSelector := mediadevices.NewCodecSelector(
			mediadevices.WithVideoEncoders(&vpxParams),
		)

		fmt.Println(mediadevices.EnumerateDevices())

		s, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
			Video: func(c *mediadevices.MediaTrackConstraints) {
				c.FrameFormat = prop.FrameFormat(frame.FormatYUY2)
				c.Width = prop.Int(640)
				c.Height = prop.Int(480)
			},
			Codec: codecSelector,
		})

		if err != nil {
			panic(err)
		}

		for _, track := range s.GetTracks() {
			track.OnEnded(func(err error) {
				fmt.Printf("Track (ID: %s) ended with error: %v\n",
					track.ID(), err)
			})
			_, err = rtc.Publish(track)
			if err != nil {
				panic(err)
			} else {
				break // only publish first track, thanks
			}
		}

		select {}
	}
}
