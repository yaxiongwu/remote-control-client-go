//go:build !js
// +build !js

package main

/*
#cgo pkg-config: opus
#include <opus.h>

int
bridge_decoder_get_last_packet_duration(OpusDecoder *st, opus_int32 *samples)
{
	return opus_decoder_ctl(st, OPUS_GET_LAST_PACKET_DURATION(samples));
}
*/
import "C"
import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"unsafe"

	log "github.com/pion/ion-log"
	"github.com/pion/rtp"

	sdk "github.com/yaxiongwu/remote-control-client-go"
	//ilog "github.com/pion/ion-log"
	"github.com/pion/mediadevices"
	"github.com/pion/webrtc/v3"

	// Note: If you don't have a camera or microphone or your adapters are not supported,
	//       you can always swap your adapters with our dummy adapters below.
	// _ "github.com/pion/mediadevices/pkg/driver/videotest"
	// _ "github.com/pion/mediadevices/pkg/driver/audiotest"

	//"github.com/pion/mediadevices/pkg/codec/x264"
	//"github.com/pion/mediadevices/pkg/codec/vpx"

	"github.com/hajimehoshi/oto/v2"
	"github.com/pion/mediadevices/pkg/codec/mmal"
	//"github.com/pion/mediadevices/pkg/codec/opus"
	_ "github.com/pion/mediadevices/pkg/driver/camera"     // This is required to register camera adapter
	_ "github.com/pion/mediadevices/pkg/driver/microphone" // This is required to register microphone adapter
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/prop"
	gst "github.com/yaxiongwu/remote-control-client-go/pkg/gstreamer-src"
)

var (
//log = ilog.NewLoggerWithFields(ilog.DebugLevel, "main", nil)
)

func main() {

	// parse flag
	var session, addr string

	//flag.StringVar(&addr, "addr", "192.168.1.199:5551", "ion-sfu grpc addr")
	flag.StringVar(&addr, "addr", "120.78.200.246:5551", "ion-sfu grpc addr")
	flag.StringVar(&session, "session", "ion", "join session name")
	flag.Parse()
	
	audioSrc := " autoaudiosrc ! audio/x-raw"
	
	//videoSrc := flag.String("video-src", "videotestsrc", "GStreamer video src")
	flag.Parse()

	// Create a audio track
	audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio", "pion1")
	if err != nil {
		panic(err)
	}

	//log.SetFlags(log.Ldate | log.Lshortfile)

	//在树莓派上控制时开启

	speed := make(chan int)
	pi := sdk.Init(26, 19, 13, 6)
	pi.SpeedControl(speed)

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
		})
	}

	/*使用树莓派时切换编码器*/
	mmalParams, err := mmal.NewParams()
	if err != nil {
		panic(err)
	}
	mmalParams.BitRate = 1_500_000 // 500kbps
	//opusParams, err := opus.NewParams()
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(opusParams)

	codecSelector := mediadevices.NewCodecSelector(
		mediadevices.WithVideoEncoders(&mmalParams),
		//mediadevices.WithAudioEncoders(&opusParams),
	)

	/*

		x264Params, _ := x264.NewParams()
		x264Params.Preset = x264.PresetMedium
		x264Params.BitRate = 6_000_000 // 1mbpsvs

		codecSelector := mediadevices.NewCodecSelector(
			mediadevices.WithVideoEncoders(&x264Params),
		)
	*/
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
		//Audio: func(c *mediadevices.MediaTrackConstraints) {},
		Codec: codecSelector,
	})

	if err != nil {
		fmt.Println("mediadevices.GetUserMedia err:")
		panic(err)
	}

	rtc.OnTrack = func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		codec := track.Codec()
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

			decoder, err := NewOpusDecoder(samplingRate, numOfChannels)
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
              tracks = append(tracks, audioTrack)
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
			gst.CreatePipeline("opus", []*webrtc.TrackLocalStaticSample{audioTrack}, audioSrc).Start()
		} else if state == webrtc.ICEConnectionStateDisconnected {

			log.Infof("sub ICEConnectionStateDisconnected")
		}
	}
	select {}
}

var errDecUninitialized = fmt.Errorf("opus decoder uninitialized")

type Decoder struct {
	p *C.struct_OpusDecoder
	// Same purpose as encoder struct
	mem         []byte
	sample_rate int
	channels    int
	opus_data   []byte
}

// NewDecoder allocates a new Opus decoder and initializes it with the
// appropriate parameters. All related memory is managed by the Go GC.
func NewOpusDecoder(sample_rate int, channels int) (*Decoder, error) {
	var dec Decoder
	err := dec.Init(sample_rate, channels)
	if err != nil {
		return nil, err
	}
	return &dec, nil
}

func (dec *Decoder) Init(sample_rate int, channels int) error {
	if dec.p != nil {
		return fmt.Errorf("opus decoder already initialized")
	}
	if channels != 1 && channels != 2 {
		return fmt.Errorf("Number of channels must be 1 or 2: %d", channels)
	}
	size := C.opus_decoder_get_size(C.int(channels))
	dec.sample_rate = sample_rate
	dec.channels = channels
	dec.mem = make([]byte, size)
	fmt.Println("decode init size:", size)
	dec.p = (*C.OpusDecoder)(unsafe.Pointer(&dec.mem[0]))
	errno := C.opus_decoder_init(
		dec.p,
		C.opus_int32(sample_rate),
		C.int(channels))
	if errno != 0 {
		return errors.New("errno")
	}
	return nil
}
func (dec *Decoder) SetOpusData(data []byte) error {
	dec.opus_data = data // *(*[]byte)(unsafe.Pointer(&data))
	return nil
}

//这里做一个fifo，wirte在on.track中调用，read在play中调用
func (dec *Decoder) Read(pcm []byte) (int, error) {
	if dec.p == nil {
		return 0, errDecUninitialized
	}
	//fmt.Println("2:", len(dec.opus_data)) //, &dec.opus_data)
	if len(dec.opus_data) == 0 {
		return 0, fmt.Errorf("opus: no data supplied")
	}
	if len(pcm) == 0 {
		return 0, fmt.Errorf("opus: target buffer empty")
	}
	if cap(pcm)%dec.channels != 0 {
		return 0, fmt.Errorf("opus: target buffer capacity must be multiple of channels")
	}
	n := int(C.opus_decode(
		dec.p,
		(*C.uchar)(&dec.opus_data[0]),
		C.opus_int32(len(dec.opus_data)),
		(*C.opus_int16)((*int16)(unsafe.Pointer(&pcm[0]))),
		C.int((cap(pcm)/dec.channels)/2),
		0))
	if n < 0 {
		return 0, errors.New("n<0")
	}
	return n * 2, nil
}

func (dec *Decoder) Write(pcm []byte) (int, error) {
	dec.opus_data = pcm
	length := len(dec.opus_data)
	//fmt.Printf("lenght:%d\n", length)
	return length, nil
}
