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
	"net"
	"unsafe"

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

	var session, addr string
	//flag.StringVar(&addr, "addr", "192.168.1.199:5551", "ion-sfu grpc addr")
	flag.StringVar(&addr, "addr", "120.78.200.246:5551", "ion-sfu grpc addr")
	flag.StringVar(&session, "session", "ion", "join session name")
	//audioSrc := " autoaudiosrc ! audio/x-raw"
	//omxh264enc可能需要设置长宽为16倍整数，否则会出现"green band"，一道偏色栏
	videoSrc := " autovideosrc ! video/x-raw, width=880,height=720 ! videoconvert ! queue"
	//videoSrc := flag.String("video-src", "videotestsrc", "GStreamer video src")
	flag.Parse()

	// Create a audio track
	// audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio", "pion1")
	// if err != nil {
	// 	panic(err)
	// }

	// Create a video track
	//videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "video/vp8"}, "video", "pion2")
	videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "video/H264"}, "video", "pion2")
	if err != nil {
		panic(err)
	}

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
		})
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
		// if state == webrtc.ICEConnectionStateDisconnected {
		// 	rtc.GetSubTransport().GetPeerConnection().Close()
		// 	log.Infof("rtc.GetSubTransport().GetPeerConnection().Close()")
		// }
		if state == webrtc.ICEConnectionStateConnected {
			//var tracks = [...]webrtc.TrackLocal{}

			_, err = rtc.Publish(videoTrack) //, audioTrack)
			//gst.CreatePipeline("vp8", []*webrtc.TrackLocalStaticSample{videoTrack}, videoSrc).Start()
			gst.CreatePipeline("h264", []*webrtc.TrackLocalStaticSample{videoTrack}, videoSrc).Start()
			//gst.CreatePipeline("opus", []*webrtc.TrackLocalStaticSample{audioTrack}, audioSrc).Start()
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
