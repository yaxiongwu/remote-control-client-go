//go:build !js
// +build !js

package main

import (
	"flag"
	"fmt"
	"net"

	log "github.com/pion/ion-log"
	"github.com/pion/rtp"

	sdk "github.com/yaxiongwu/remote-control-client-go"
	//ilog "github.com/pion/ion-log"

	"github.com/pion/webrtc/v3"

	// Note: If you don't have a camera or microphone or your adapters are not supported,
	//       you can always swap your adapters with our dummy adapters below.
	// _ "github.com/pion/mediadevices/pkg/driver/videotest"
	// _ "github.com/pion/mediadevices/pkg/driver/audiotest"
	//"github.com/pion/mediadevices/pkg/codec/mmal"

	// This is required to use opus audio encoder

	//"github.com/pion/mediadevices/pkg/codec/mmal"
	//"github.com/pion/mediadevices/pkg/codec/vpx"
	"github.com/hajimehoshi/oto/v2"
	_ "github.com/pion/mediadevices/pkg/driver/camera"     // This is required to register camera adapter
	_ "github.com/pion/mediadevices/pkg/driver/microphone" // This is required to register microphone adapter
	gst "github.com/yaxiongwu/remote-control-client-go/pkg/gstreamer-src"
	opusdecoder "github.com/yaxiongwu/remote-control-client-go/pkg/opus/decoder"
	"github.com/yaxiongwu/remote-control-client-go/pkg/rtmpudp"
)

type udpConn struct {
	conn        *net.UDPConn
	port        int
	payloadType uint8
}

var (
// log = ilog.NewLoggerWithFields(ilog.DebugLevel, "main", nil)
)

func main() {

	var session, addr string
	//flag.StringVar(&addr, "addr", "192.168.1.199:5551", "ion-sfu grpc addr")
	flag.StringVar(&addr, "addr", "120.78.200.246:5551", "ion-sfu grpc addr")
	flag.StringVar(&session, "session", "ion", "join session name")
	audioSrc := " autoaudiosrc ! audio/x-raw"
	//omxh264enc可能需要设置长宽为32倍整数，否则会出现"green band"，一道偏色栏
	//videoSrc := " autovideosrc ! video/x-raw, width=640, height=480 ! videoconvert ! queue"
	videoSrc := " autovideosrc ! video/x-raw, width=600, height=400 ! videoconvert ! queue"
	//videoSrc := flag.String("video-src", "videotestsrc", "GStreamer video src")
	flag.Parse()
	// Create a video track
	//videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "video/vp8"}, "video", "pion2")
	videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "video/H264"}, "video", "pion2")
	if err != nil {
		panic(err)
	}
	// Create a audio track
	audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio", "pion1")
	if err != nil {
		panic(err)
	}

	rtmpudp := rtmpudp.Init("5000")
	//gst.CreatePipeline("vp8", []*webrtc.TrackLocalStaticSample{videoTrack}, videoSrc).Start()
	gst.CreatePipeline("h264", []*webrtc.TrackLocalStaticSample{videoTrack}, videoSrc, rtmpudp.GetConn()).Start()
	gst.CreatePipeline("opus", []*webrtc.TrackLocalStaticSample{audioTrack}, audioSrc, rtmpudp.GetConn()).Start()

	connector := sdk.NewConnector(addr)
	rtc, err := sdk.NewRTC(connector)
	if err != nil {
		panic(err)
	}

	rtc.OnPubIceConnectionStateChange = func(state webrtc.ICEConnectionState) {
		log.Infof("Pub Connection state changed: %s", state)
		if state == webrtc.ICEConnectionStateDisconnected || state == webrtc.ICEConnectionStateFailed {
			log.Infof("rtc.GetPubTransport().GetPeerConnection().Close()")
			rtc.ReStart()
		}
	}

	log.Infof("rtc.GetSubTransport():%v,rtc.GetSubTransport().GetPeerConnection():%v", rtc.GetSubTransport(), rtc.GetSubTransport().GetPeerConnection())

	err = rtc.RegisterNewVideoSource("ion", "PiVideoSource")

	rtc.OnDataChannel = func(dc *webrtc.DataChannel) {

		log.Infof("rtc.OnDatachannel:%v", dc.Label())
		dc.OnOpen(func() {
			log.Infof("%v,dc.onopen,dc.ReadyState:%v", dc.Label(), dc.ReadyState())
			//	dc.SendText("wuyaxiong nbcl")
		})

	}

	rtc.OnTrack = func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		codec := track.Codec()
		log.Infof("track.Codec():%v", codec)
		if codec.MimeType == "audio/opus" {
			samplingRate := 48000

			// Number of channels (aka locations) to play sounds from. Either 1 or 2.
			// 1 is mono sound, and 2 is stereo (most speakers are stereo).
			numOfChannels := 1

			// Bytes used by a channel to represent one sample. Either 1 or 2 (usually 2).
			audioBitDepth := 2

			otoCtx, readyChan, err := oto.NewContext(samplingRate, numOfChannels, audioBitDepth)
			if err != nil {
				panic("oto.NewContext failed: " + err.Error())
			}
			// It might take a bit for the hardware audio devices to be ready, so we wait on the channel.
			<-readyChan

			decoder, err := opusdecoder.NewOpusDecoder(samplingRate, numOfChannels)
			if err != nil {
				fmt.Printf("Error creating")
			}
			player := otoCtx.NewPlayer(decoder)
			defer player.Close()
			//player.Play()
			//pipeReader, pipeWriter := io.Pipe()

			b := make([]byte, 1500)
			rtpPacket := &rtp.Packet{}
			for {

				// Read
				n, _, readErr := track.Read(b)
				if readErr != nil {
					log.Errorf("OnTrack read error: %v", readErr)
					return
					//panic(readErr)
				}

				// Unmarshal the packet and update the PayloadType
				if err = rtpPacket.Unmarshal(b[:n]); err != nil {
					log.Errorf("OnTrack UnMarshal error: %v", err)
					return
					//panic(err)
				}

				//复制一份，以防覆盖
				temp := make([]byte, len(rtpPacket.Payload))
				copy(temp, rtpPacket.Payload)
				//decoder.SetOpusData(rtpPacket.Payload)
				decoder.Write(temp)

				player.Play()

			}
		}
	}

	rtc.OnSubIceConnectionStateChange = func(state webrtc.ICEConnectionState) {
		log.Infof("Sub Connection state changed: %s", state)
		// if state == webrtc.ICEConnectionStateDisconnected {
		// 	rtc.GetSubTransport().GetPeerConnection().Close()
		// 	log.Infof("rtc.GetSubTransport().GetPeerConnection().Close()")
		// }
		if state == webrtc.ICEConnectionStateConnected {
			//var tracks = [...]webrtc.TrackLocal{}

			_, err = rtc.Publish(videoTrack, audioTrack)

			if err != nil {
				log.Errorf("join err=%v", err)
				panic(err)
			}
		} else if state == webrtc.ICEConnectionStateDisconnected {

			log.Infof("sub ICEConnectionStateDisconnected")
		}
	}

	select {}
}
