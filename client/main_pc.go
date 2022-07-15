package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"

	log "github.com/pion/ion-log"
	"github.com/pion/rtp"

	sdk "github.com/YaxiongWu/remote-control-client-go"
	//ilog "github.com/pion/ion-log"
	"github.com/pion/mediadevices"
	"github.com/pion/webrtc/v3"

	// Note: If you don't have a camera or microphone or your adapters are not supported,
	//       you can always swap your adapters with our dummy adapters below.
	// _ "github.com/pion/mediadevices/pkg/driver/videotest"
	// _ "github.com/pion/mediadevices/pkg/driver/audiotest"
	//"github.com/pion/mediadevices/pkg/codec/mmal"

	"github.com/pion/mediadevices/pkg/codec/opus" // This is required to use opus audio encoder
	"github.com/pion/mediadevices/pkg/codec/x264"

	//"github.com/pion/mediadevices/pkg/codec/mmal"
	//"github.com/pion/mediadevices/pkg/codec/vpx"
	_ "github.com/pion/mediadevices/pkg/driver/camera"     // This is required to register camera adapter
	_ "github.com/pion/mediadevices/pkg/driver/microphone" // This is required to register microphone adapter
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/prop"
)

type udpConn struct {
	conn        *net.UDPConn
	port        int
	payloadType uint8
}

var (
//log = ilog.NewLoggerWithFields(ilog.DebugLevel, "main", nil)
)

func main() {

	// parse flag
	var session, addr string
	//var rtpSenders []*webrtc.RTPSender

	//meidaOpen := false
	//flag.StringVar(&addr, "addr", "192.168.1.199:5551", "ion-sfu grpc addr")
	flag.StringVar(&addr, "addr", "120.78.200.246:5551", "ion-sfu grpc addr")
	flag.StringVar(&session, "session", "ion", "join session name")
	flag.Parse()
	subConnectioned := true

	//log.SetFlags(log.Ldate | log.Lshortfile)

	//在树莓派上控制时开启
	/*
		speed := make(chan int)
		pi := sdk.Init(26, 19, 13, 6)
		pi.SpeedControl(speed)
	*/

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
			subConnectioned = false
			rtc.ReStart()
		}
		log.Infof("Pub Connection state changed: %s", state)
	}

	log.Infof("rtc.GetSubTransport():%v,rtc.GetSubTransport().GetPeerConnection():%v", rtc.GetSubTransport(), rtc.GetSubTransport().GetPeerConnection())

	err = rtc.RegisterNewVideoSource("ion", "PiVideoSource")

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
			/*使用树莓派时开启*/
			/*
					//if recvData["type"] != nil {
					switch recvData["type"] {
					case 1: //方向
						//每次方向摇杆放开就会回到(0,0)，如果y=0，固定为往前走，这样会导致永远不会往后走
						pi.DirectionControl(recvData["x"], recvData["y"])
					case 2: //速度
						//if(recvData["speed"]!=nil){
						speed <- recvData["speed"]
						//}
					} //switch
					//} //if
				 log.Infof("recvData:%v,%v", recvData["t"], recvData["x"])
			*/
		})
	}

	/*使用树莓派时切换编码器*/
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

	x264Params, _ := x264.NewParams()
	x264Params.Preset = x264.PresetMedium
	x264Params.BitRate = 6_000_000 // 1mbpsvs

	opusParams, err := opus.NewParams()
	if err != nil {
		panic(err)
	}
	fmt.Println(opusParams)
	codecSelector := mediadevices.NewCodecSelector(
		mediadevices.WithVideoEncoders(&x264Params),
		mediadevices.WithAudioEncoders(&opusParams),
	)

	log.Infof("mediadevices.EnumerateDevices():")
	fmt.Println(mediadevices.EnumerateDevices())

	s, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Video: func(c *mediadevices.MediaTrackConstraints) {
			c.FrameFormat = prop.FrameFormat(frame.FormatYUY2)
			c.Width = prop.Int(800)
			c.Height = prop.Int(600)
		},
		Audio: func(c *mediadevices.MediaTrackConstraints) {},
		Codec: codecSelector,
	})

	if err != nil {
		log.Errorf("mediadevices.GetUserMedia err:")
		panic(err)
	}

	for _, track1 := range s.GetTracks() {
		log.Infof("Track (ID: %s) : %v\n", track1.ID())
	}
	// Create a local addr
	var laddr *net.UDPAddr
	if laddr, err = net.ResolveUDPAddr("udp", "127.0.0.1:"); err != nil {
		panic(err)
	}

	// Prepare udp conns
	// Also update incoming packets with expected PayloadType, the browser may use
	// a different value. We have to modify so our stream matches what rtp-forwarder.sdp expects
	udpConns := map[string]*udpConn{
		"audio": {port: 4000, payloadType: 111},
		"video": {port: 4002, payloadType: 96},
	}

	for _, c := range udpConns {
		// Create remote addr
		var raddr *net.UDPAddr
		if raddr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", c.port)); err != nil {
			panic(err)
		}

		// Dial udp
		if c.conn, err = net.DialUDP("udp", laddr, raddr); err != nil {
			panic(err)
		}
		defer func(conn net.PacketConn) {
			if closeErr := conn.Close(); closeErr != nil {
				panic(closeErr)
			}
		}(c.conn)
	}

	rtc.OnTrack = func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		log.Infof("[S=>C] got track streamId=%v kind=%v ssrc=%v ", track.StreamID(), track.Kind(), track.SSRC())
		c, ok := udpConns[track.Kind().String()]
		if !ok {
			return
		}

		// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
		// go func() {
		// 	ticker := time.NewTicker(time.Second * 2)
		// 	for range ticker.C {
		// 		if rtcpErr := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(track.SSRC())}}); rtcpErr != nil {
		// 			fmt.Println(rtcpErr)
		// 		}
		// 	}
		// }()

		b := make([]byte, 1500)
		rtpPacket := &rtp.Packet{}
		for {
			// if rtc.GetSubTransport().GetPeerConnection().ICEConnectionState() != webrtc.ICEConnectionStateConnected {
			// 	return
			// }
			// Read
			if !subConnectioned {
				log.Infof("subConnectioned:%v", subConnectioned)
				return
			}

			n, _, readErr := track.Read(b)
			if readErr != nil {
				return
				log.Errorf("readErr:%s", readErr)
				panic(readErr)
			}

			// Unmarshal the packet and update the PayloadType
			if err = rtpPacket.Unmarshal(b[:n]); err != nil {
				log.Errorf("err:%s", err)
				panic(err)
			}
			rtpPacket.PayloadType = c.payloadType

			// Marshal into original buffer with updated PayloadType
			if n, err = rtpPacket.MarshalTo(b); err != nil {
				log.Errorf("err:%s", err)
				panic(err)
			}

			// Write

			if _, writeErr := c.conn.Write(b[:n]); writeErr != nil {
				// For this particular example, third party applications usually timeout after a short
				// amount of time during which the user doesn't have enough time to provide the answer
				// to the browser.
				// That's why, for this particular example, the user first needs to provide the answer
				// to the browser then open the third party application. Therefore we must not kill
				// the forward on "connection refused" errors
				var opError *net.OpError
				if errors.As(writeErr, &opError) && opError.Err.Error() == "write: connection refused" {
					continue
				}
				//log.Errorf("err:%s", err)
				panic(err)
			}

		}
		// codec := track.Codec()
		// if codec.MimeType == "audio/opus" {
		// 	fmt.Println("Got Opus track, saving to disk as output.ogg,clockRate:%v,channels:%v", codec.ClockRate, codec.Channels)
		// 	i, oggNewErr := oggwriter.New("output.ogg", codec.ClockRate, codec.Channels)
		// 	if oggNewErr != nil {
		// 		panic(oggNewErr)
		// 	}
		// 	saveToDisk(i, track)
		// } else if codec.MimeType == "video/VP8" {
		// 	fmt.Println("Got VP8 track, saving to disk as output.ivf")
		// 	i, ivfNewErr := ivfwriter.New("output.ivf")
		// 	if ivfNewErr != nil {
		// 		panic(ivfNewErr)
		// 	}
		// 	saveToDisk(i, track)
		// }

	}
	rtc.OnSubIceConnectionStateChange = func(state webrtc.ICEConnectionState) {
		log.Infof("Sub Connection state changed: %s", state)
		// if state == webrtc.ICEConnectionStateDisconnected {
		// 	rtc.GetSubTransport().GetPeerConnection().Close()
		// 	log.Infof("rtc.GetSubTransport().GetPeerConnection().Close()")
		// }
		if state == webrtc.ICEConnectionStateConnected {
			//var tracks = [...]webrtc.TrackLocal{}
			tracks := []webrtc.TrackLocal{}
			for _, track := range s.GetTracks() {
				track.OnEnded(func(err error) {
					log.Infof("Track (ID: %s) ended with error: %v\n", track.ID(), err)
				})
				tracks = append(tracks, track)

				//tracks[index] = Track
				//rtpSenders, err = rtc.Publish(track)
				//if err != nil {
				//	panic(err)
				//} else {
				//meidaOpen = true
				//break // only publish first track, thanks
				//}
			}

			_, err = rtc.Publish(tracks...)
			if err != nil {
				panic(err)
			} else {
				//meidaOpen = true
				//break // only publish first track, thanks
			}

			if err != nil {
				log.Errorf("join err=%v", err)
				panic(err)
			}
		} else if state == webrtc.ICEConnectionStateDisconnected {
			subConnectioned = false
			log.Infof("subConnectioned: %v", subConnectioned)
		}
	}
	select {}
}
