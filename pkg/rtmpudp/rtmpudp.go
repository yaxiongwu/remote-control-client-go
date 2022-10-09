package rtmpudp

import "C"
import (
	"fmt"
	"net"
	"os/exec"
)

type RtmpUdp struct {
	conn *net.UDPConn
	//port   string
	Buffer       [64000]byte
	BufferLength int
	haveNewData  chan bool
}

func Init(port string) *RtmpUdp {
	cmd := exec.Command("bash", "-c", "gst-launch-1.0 udpsrc port="+port+" ! queue ! h264parse ! flvmux ! rtmpsink location='rtmp://live-push.bilivideo.com/live-bvc/?streamname=live_443203481_72219565&key=0c399147659bfa24be5454360c227c21&schedule=rtmp&pflag=1'")
	err := cmd.Start()
	if err != nil {
		fmt.Printf("gst-launch udp error:%s\n", err)
	}

	addrRtmp, err2 := net.ResolveUDPAddr("udp", "localhost:"+port)
	if err2 != nil {
		fmt.Printf("net.ResolveUDPAddr %s ", err2)
	}
	conn, err3 := net.DialUDP("udp", nil, addrRtmp)
	if err3 != nil {
		fmt.Printf("net.DialUDP %s ", err3)
	}
	rtmpudp := &RtmpUdp{
		conn: conn,
		//port: port,
		haveNewData: make(chan bool),
		//buf:         make([]byte, 10000),
	}
	//go WirteUdpData(rtmpudp.haveNewData, rtmpudp.buffer)
	go rtmpudp.wirteUdpDataChannel()
	return rtmpudp
}

func (r *RtmpUdp) GetConn() *net.UDPConn {
	return r.conn
}

func (r *RtmpUdp) wirteUdpDataChannel() error {
	//fmt.Printf("WirteUdpDataChannel before r.haveNewData:%v\n\n\n\n", r.haveNewData)
	//r.haveNewData <- true

	for {
		select {
		case <-r.haveNewData:
			_, err := r.conn.Write(r.Buffer[:r.BufferLength])
			if err != nil {
				return err
			}
			//fmt.Printf("WirteUdpDataChannel after  <-r.HaveNewData\n\n\n")
			//return nil
		}
	}
}

//func (r *RtmpUdp) CopyUdpData(buffer unsafe.Pointer, bufferLen C.int) error {
func (r *RtmpUdp) CopyUdpData() error {
	//fmt.Printf("rtmpudp.go CopyUdpData()  bufferLen:%v\n\n\n\n", bufferLen)
	//r.haveNewData <- true
	//len := int(bufferLen)
	//copy(r.buf[0:len], C.GoBytes(buffer, bufferLen)[0:bufferLen])
	//r.bufLength = len
	//r.buf = data
	r.haveNewData <- true
	return nil
}
