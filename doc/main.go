package main

import (
	"encoding/json"
	"flag"
	"fmt"

	sdk "github.com/YaxiongWu/remote-control-client-go"
	ilog "github.com/pion/ion-log"
	"github.com/pion/mediadevices"
	"github.com/pion/webrtc/v3"

	// Note: If you don't have a camera or microphone or your adapters are not supported,
	//       you can always swap your adapters with our dummy adapters below.
	// _ "github.com/pion/mediadevices/pkg/driver/videotest"
	// _ "github.com/pion/mediadevices/pkg/driver/audiotest"
	"github.com/pion/mediadevices/pkg/codec/x264"
	//"github.com/pion/mediadevices/pkg/codec/mmal"
	//"github.com/pion/mediadevices/pkg/codec/vpx"
	_ "github.com/pion/mediadevices/pkg/driver/camera"     // This is required to register camera adapter
	_ "github.com/pion/mediadevices/pkg/driver/microphone" // This is required to register microphone adapter
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/prop"
)

var (
	log = ilog.NewLoggerWithFields(ilog.DebugLevel, "main", nil)
)

func main() {

	// parse flag
	var session, addr string
	var rtpSenders []*webrtc.RTPSender
	meidaOpen := false
	//flag.StringVar(&addr, "addr", "192.168.1.199:5551", "ion-sfu grpc addr")
	flag.StringVar(&addr, "addr", "120.78.200.246:5551", "ion-sfu grpc addr")
	flag.StringVar(&session, "session", "ion", "join session name")
	flag.Parse()

	//在树莓派上控制时开启
	//speed := make(chan int)
	//pi := sdk.Init(26, 19, 13, 6)
	//pi.SpeedControl(speed)

	connector := sdk.NewConnector(addr)
	rtc, err := sdk.NewRTC(connector)
	if err != nil {
		panic(err)
	}

	rtc.OnPubIceConnectionStateChange = func(state webrtc.ICEConnectionState) {
		if state == webrtc.ICEConnectionStateDisconnected {
			// for _, rtpSend := range rtpSenders {
			// 	rtc.GetPubTransport().GetPeerConnection().RemoveTrack(rtpSend)
			// }
			// rtc.UnPublish(rtpSenders)
			log.Infof("rtc.GetPubTransport().GetPeerConnection().Close()")
			rtc.ReStart()
		}
		log.Infof("Pub Connection state changed: %s", state)
	}
	log.Infof("rtc.GetSubTransport():%v,rtc.GetSubTransport().GetPeerConnection():%v", rtc.GetSubTransport(), rtc.GetSubTransport().GetPeerConnection())

	err = rtc.CreateSession("ion", "PiVideoSource")
	// var infos []*sdk.Subscription
	// infos = append(infos, &sdk.Subscription{
	// 	TrackId: "ion",
	// 	Layer:   "PiVideoSource",
	// })

	// err = rtc.Subscribe(infos)
	//err = rtc.Join(session, "videoSource"+sdk.RandomKey(8))

	// dataChanenl, err := rtc.GetSubTransport().GetPeerConnection().CreateDataChannel("test-channel", &webrtc.DataChannelInit{})
	// log.Infof("dataChannel label:%v,ReadyState:%v", dataChanenl.Label(), dataChanenl.ReadyState())
	// dataChanenl.OnOpen(func() {
	// 	log.Infof("OnOpen dataChannel label:%v,ReadyState:%v", dataChanenl.Label(), dataChanenl.ReadyState())
	// 	dataChanenl.SendText("wuyaxiong test,onopen()")
	// })
	// dataChanenl.OnMessage(func(msg webrtc.DataChannelMessage) {
	// 	log.Infof("get msg from:%v,msg:%v", dataChanenl.Label(), msg)
	// })

	rtc.OnDataChannel = func(dc *webrtc.DataChannel) {
		recvData := make(map[string]int)
		log.Infof("rtc.OnDatachannel:%v", dc.Label())
		dc.OnOpen(func() {
			log.Infof("%v,dc.onopen,dc.ReadyState:%v", dc.Label(), dc.ReadyState())
			//	dc.SendText("wuyaxiong nbcl")
		})

		dc.OnMessage(func(msg webrtc.DataChannelMessage) {
			log.Infof("get msg from:%v,msg:%s", dc.Label(), msg.Data)
			err := json.Unmarshal(msg.Data, &recvData)
			if err != nil {
				log.Errorf("Unmarshal:err %v", err)
				return
			}

			//if(recvData["type"] != nil){
			switch recvData["type"] {
			case 1: //方向
				//每次方向摇杆放开就会回到(0,0)，如果y=0，固定为往前走，这样会导致永远不会往后走
				//if recvData["y"] == 0 {
				//    break
				//}
				//   if(recvData["x"] != nil){
				//   dx <- recvData["x"]
				//}
				// if(recvData.y != nil){
				//  dy <- recvData["y"]
				//}
				//	pi.DirectionControl(recvData["x"], recvData["y"])
			case 2: //速度
				//if(recvData["speed"]!=nil){
				//speed <- recvData["speed"]
				//}
			} //switch
			//}//if
			log.Infof("recvData:%v,%v", recvData["t"], recvData["x"])
		})
	}
	/*
			 mmalParams, err := mmal.NewParams()
			 if err != nil {
			  	panic(err)
			   }
			 mmalParams.BitRate = 1_500_000 // 500kbps
		     codecSelector := mediadevices.NewCodecSelector(
				      mediadevices.WithVideoEncoders(&mmalParams),
		     )
	*/

	// 	log.Infof("c.OnDataChannel = func,dc.ReadyState:%v", dc.ReadyState())

	x264Params, _ := x264.NewParams()
	x264Params.Preset = x264.PresetMedium
	x264Params.BitRate = 3_000_000 // 1mbpsvs

	codecSelector := mediadevices.NewCodecSelector(
		mediadevices.WithVideoEncoders(&x264Params),
	)

	// vpxParams, err := vpx.NewVP8Params()
	// if err != nil {
	// 	panic(err)
	// }
	// vpxParams.BitRate = 500_000 // 500kbps

	// codecSelector := mediadevices.NewCodecSelector(
	// 	mediadevices.WithVideoEncoders(&vpxParams),
	// )

	fmt.Println(mediadevices.EnumerateDevices())

	s, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Video: func(c *mediadevices.MediaTrackConstraints) {
			c.FrameFormat = prop.FrameFormat(frame.FormatYUY2)
			c.Width = prop.Int(800)
			c.Height = prop.Int(600)
		},
		Codec: codecSelector,
	})

	if err != nil {
		fmt.Println("mediadevices.GetUserMedia err:")
		panic(err)
	}

	rtc.OnSubIceConnectionStateChange = func(state webrtc.ICEConnectionState) {
		log.Infof("Sub Connection state changed: %s", state)
		// if state == webrtc.ICEConnectionStateDisconnected {
		// 	rtc.GetSubTransport().GetPeerConnection().Close()
		// 	log.Infof("rtc.GetSubTransport().GetPeerConnection().Close()")
		// }
		if state == webrtc.ICEConnectionStateConnected {

			for _, track := range s.GetTracks() {
				track.OnEnded(func(err error) {
					fmt.Printf("Track (ID: %s) ended with error: %v\n",
						track.ID(), err)
				})
				rtpSenders, err = rtc.Publish(track)
				if err != nil {
					panic(err)
				} else {
					meidaOpen = true
					break // only publish first track, thanks
				}
			}

			if err != nil {
				log.Errorf("join err=%v", err)
				panic(err)
			}
		}

	}

	select {}
}
