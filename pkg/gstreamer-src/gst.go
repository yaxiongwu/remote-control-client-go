// Package gst provides an easy API to create an appsink pipeline
package gst

/*
#cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0

#include "gst.h"

*/
import "C"
import (
	"fmt"
	"net"
	"sync"
	"time"
	"unsafe"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

func init() {
	go C.gstreamer_send_start_mainloop()
}

// Pipeline is a wrapper for a GStreamer Pipeline
type Pipeline struct {
	Pipeline  *C.GstElement
	tracks    []*webrtc.TrackLocalStaticSample
	id        int
	codecName string
	clockRate float32
	conn      *net.UDPConn
}

var pipelines = make(map[int]*Pipeline)
var pipelinesLock sync.Mutex

const (
	videoClockRate = 90000
	audioClockRate = 48000
	pcmClockRate   = 8000
)

// CreatePipeline creates a GStreamer Pipeline
//pipelineSink videoSrc := " autovideosrc ! video/x-raw, width=640, height=480 ! videoconvert ! queue"
func CreatePipeline(codecName string, tracks []*webrtc.TrackLocalStaticSample, pipelineSrc string, conn *net.UDPConn) *Pipeline {
	//pipelineStr := "tee name=t ! appsink name=appsink t. ! queue ! flvmux ! rtmpsink location='rtmp://live-push.bilivideo.com/live-bvc/?streamname=live_443203481_72219565&key=0c399147659bfa24be5454360c227c21&schedule=rtmp&pflag=1'"
	//pipelineStr := "tee name=t ! appsink name=appsink t. ! queue ! videoconvert ! flvmux ! filesink location=test.flv"
	pipelineStr := "appsink name=appsink"
	var clockRate float32

	switch codecName {
	case "vp8":
		pipelineStr = pipelineSrc + " ! vp8enc error-resilient=partitions keyframe-max-dist=10 auto-alt-ref=true cpu-used=5 deadline=1 ! " + pipelineStr
		clockRate = videoClockRate

	case "vp9":
		pipelineStr = pipelineSrc + " ! vp9enc ! " + pipelineStr
		clockRate = videoClockRate

	case "h264-omxh264enc":
		//pipelineStr = "autovideosrc ! video/x-raw, width=640, height=480 ! videoconvert ! video/x-raw,format=I420 ! x264enc speed-preset=ultrafast tune=zerolatency key-int-max=10 ! tee name =t ! queue ! appsink name=appsink t. ! queue ! flvmux ! rtmpsink location='rtmp://live-push.bilivideo.com/live-bvc/?streamname=live_443203481_72219565&key=0c399147659bfa24be5454360c227c21&schedule=rtmp&pflag=1'"
		//pipelineStr = "autovideosrc ! video/x-raw, width=640, height=480 ! videoconvert ! video/x-raw,format=I420 ! x264enc key-int-max=20 ! tee name =t  ! queue ! flvmux ! rtmpsink location='rtmp://live-push.bilivideo.com/live-bvc/?streamname=live_443203481_72219565&key=0c399147659bfa24be5454360c227c21&schedule=rtmp&pflag=1'  t. ! queue ! video/x-h264,stream-format=byte-stream ! appsink name=appsink"
		//pipelineStr = "autovideosrc ! video/x-raw, width=640, height=480 ! videoconvert ! video/x-raw,format=I420 ! x264enc speed-preset=ultrafast tune=zerolatency key-int-max=20 ! tee name =t ! queue ! appsink name=appsink t. ! queue ! flvmux ! filesink location=test.flv "
		//pipelineStr = "autovideosrc ! video/x-raw,width=640, height=480 ! videoconvert ! video/x-raw,format=I420 ! tee name=t ! x264enc speed-preset=ultrafast tune=zerolatency key-int-max=20 ! video/x-h264,stream-format=byte-stream ! queue ! appsink name=appsink t. ! queue ! x264enc ! flvmux !  rtmpsink location='rtmp://live-push.bilivideo.com/live-bvc/?streamname=live_443203481_72219565&key=0c399147659bfa24be5454360c227c21&schedule=rtmp&pflag=1'"
		//pipelineStr = "autovideosrc ! video/x-raw, width=640, height=480 ! videoconvert ! video/x-raw,format=I420 ! x264enc speed-preset=ultrafast tune=zerolatency key-int-max=20 ! tee name =t ! queue ! appsink name=appsink"
		//pipelineStr = pipelineSrc + " ! video/x-raw,format=I420 ! x264enc speed-preset=ultrafast tune=zerolatency key-int-max=20 ! " + pipelineStr
		//pipelineStr = pipelineSrc + " ! video/x-raw,format=I420,framerate=40/1 ! omxh264enc control-rate=2 target-bitrate=10485760 interval-intraframes=14 periodicty-idr=2 ! " + pipelineStr
		//pipelineStr = pipelineSrc + " ! video/x-raw,format=I420 ! x264enc speed-preset=ultrafast tune=zerolatency key-int-max=20 ! video/x-h264,stream-format=byte-stream ! " + pipelineStr
		//pipelineStr = pipelineSrc + " ! video/x-raw,format=I420 ! omxh264enc speed-preset=ultrafast tune=zerolatency key-int-max=20 ! video/x-h264,stream-format=byte-stream ! " + pipelineStr
		pipelineStr = pipelineSrc + " ! video/x-raw,format=I420,framerate=40/1 ! omxh264enc entropy-mode=1 b-frames=0 interval-intraframes=2 control-rate=1 target-bitrate=8000000 ! video/x-h264,stream-format=byte-stream ! " + pipelineStr
		clockRate = videoClockRate

	case "h264-x264enc":
		//pipelineStr = "autovideosrc ! video/x-raw, width=640, height=480 ! videoconvert ! video/x-raw,format=I420 ! x264enc speed-preset=ultrafast tune=zerolatency key-int-max=10 ! tee name =t ! queue ! appsink name=appsink t. ! queue ! flvmux ! rtmpsink location='rtmp://live-push.bilivideo.com/live-bvc/?streamname=live_443203481_72219565&key=0c399147659bfa24be5454360c227c21&schedule=rtmp&pflag=1'"
		//pipelineStr = "autovideosrc ! video/x-raw, width=640, height=480 ! videoconvert ! video/x-raw,format=I420 ! x264enc key-int-max=20 ! tee name =t  ! queue ! flvmux ! rtmpsink location='rtmp://live-push.bilivideo.com/live-bvc/?streamname=live_443203481_72219565&key=0c399147659bfa24be5454360c227c21&schedule=rtmp&pflag=1'  t. ! queue ! video/x-h264,stream-format=byte-stream ! appsink name=appsink"
		//pipelineStr = "autovideosrc ! video/x-raw, width=640, height=480 ! videoconvert ! video/x-raw,format=I420 ! x264enc speed-preset=ultrafast tune=zerolatency key-int-max=20 ! tee name =t ! queue ! appsink name=appsink t. ! queue ! flvmux ! filesink location=test.flv "
		//pipelineStr = "autovideosrc ! video/x-raw,width=640, height=480 ! videoconvert ! video/x-raw,format=I420 ! tee name=t ! x264enc speed-preset=ultrafast tune=zerolatency key-int-max=20 ! video/x-h264,stream-format=byte-stream ! queue ! appsink name=appsink t. ! queue ! x264enc ! flvmux !  rtmpsink location='rtmp://live-push.bilivideo.com/live-bvc/?streamname=live_443203481_72219565&key=0c399147659bfa24be5454360c227c21&schedule=rtmp&pflag=1'"
		//pipelineStr = "autovideosrc ! video/x-raw, width=640, height=480 ! videoconvert ! video/x-raw,format=I420 ! x264enc speed-preset=ultrafast tune=zerolatency key-int-max=20 ! tee name =t ! queue ! appsink name=appsink"
		//pipelineStr = pipelineSrc + " ! video/x-raw,format=I420,framerate=40/1 ! omxh264enc control-rate=2 target-bitrate=10485760 interval-intraframes=14 periodicty-idr=2 ! " + pipelineStr
		//pipelineStr = pipelineSrc + " ! video/x-raw,format=I420 ! x264enc speed-preset=ultrafast tune=zerolatency key-int-max=20 ! video/x-h264,stream-format=byte-stream ! " + pipelineStr
		pipelineStr = pipelineSrc + " ! video/x-raw,format=I420 ! x264enc speed-preset=ultrafast tune=zerolatency key-int-max=20 ! " + pipelineStr

		clockRate = videoClockRate

	case "opus":
		pipelineStr = pipelineSrc + " ! opusenc ! " + pipelineStr
		clockRate = audioClockRate

	case "g722":
		pipelineStr = pipelineSrc + " ! avenc_g722 ! " + pipelineStr
		clockRate = audioClockRate

	case "pcmu":
		pipelineStr = pipelineSrc + " ! audio/x-raw, rate=8000 ! mulawenc ! " + pipelineStr
		clockRate = pcmClockRate

	case "pcma":
		pipelineStr = pipelineSrc + " ! audio/x-raw, rate=8000 ! alawenc ! " + pipelineStr
		clockRate = pcmClockRate

	default:
		panic("Unhandled codec " + codecName)
	}

	pipelineStrUnsafe := C.CString(pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	pipelinesLock.Lock()
	defer pipelinesLock.Unlock()

	pipeline := &Pipeline{
		Pipeline:  C.gstreamer_send_create_pipeline(pipelineStrUnsafe),
		tracks:    tracks,
		id:        len(pipelines),
		codecName: codecName,
		clockRate: clockRate,
		conn:      conn,
	}

	pipelines[pipeline.id] = pipeline
	return pipeline
}

// Start starts the GStreamer Pipeline
func (p *Pipeline) Start() {
	C.gstreamer_send_start_pipeline(p.Pipeline, C.int(p.id))
}

// Stop stops the GStreamer Pipeline
func (p *Pipeline) Stop() {
	C.gstreamer_send_stop_pipeline(p.Pipeline)
}

//export goHandlePipelineBuffer
func goHandlePipelineBuffer(buffer unsafe.Pointer, bufferLen C.int, duration C.int, pipelineID C.int) {
	pipelinesLock.Lock()
	pipeline, ok := pipelines[int(pipelineID)]
	pipelinesLock.Unlock()
	//fmt.Printf("goHandlePipelineBuffer %d ", pipelineID)
	//buff := make([]byte, 1000)

	if ok {
		for _, t := range pipeline.tracks {
			if err := t.WriteSample(media.Sample{Data: C.GoBytes(buffer, bufferLen), Duration: time.Duration(duration)}); err != nil {
				panic(err)
			}

			if pipeline.codecName == "h264" {

				go func(conn *net.UDPConn, buffer unsafe.Pointer, bufferLen C.int) {
					/*defer func(){
					if err := recover(); err!=nil{
						fmt.Println("conn recover err:",err)
						}
					}()
					* */
					_, err1 := conn.Write(C.GoBytes(buffer, bufferLen)[0:bufferLen])
					if err1 != nil {
						panic(err1)
					}
				}(pipeline.conn, buffer, bufferLen)
			} //if pipeline.codecName == "h264" {
		} //for
	} else {
		fmt.Printf("discarding buffer, no pipeline with id %d", int(pipelineID))
	}
	C.free(buffer)
}

//gst-launch-1.0 udpsrc port=5000 ! queue ! h264parse ! flvmux ! rtmpsink location='rtmp://live-push.bilivideo.com/live-bvc/?streamname=live_443203481_72219565&key=0c399147659bfa24be5454360c227c21&schedule=rtmp&pflag=1'
//gst-launch-1.0 autovideosrc ! video/x-raw, width=880,height=720 ! videoconvert ! omxh264enc control-rate=variable target-bitrate=6000000 ! h264parse ! flvmux ! rtmpsink location='rtmp://live-push.bilivideo.com/live-bvc/?streamname=live_443203481_72219565&key=0c399147659bfa24be5454360c227c21&schedule=rtmp&pflag=1'
