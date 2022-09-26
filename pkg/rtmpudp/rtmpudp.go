package rtmpudp


import (
	
	"fmt"
	"net"
	"os/exec"

)

type RtmpUdp struct {
	conn        *net.UDPConn
	port        string
}

func Init(port string) *RtmpUdp{
    cmd :=exec.Command("bash","-c","gst-launch-1.0 udpsrc port="+port+" ! queue ! h264parse ! flvmux ! rtmpsink location='rtmp://live-push.bilivideo.com/live-bvc/?streamname=live_443203481_72219565&key=0c399147659bfa24be5454360c227c21&schedule=rtmp&pflag=1'")
	err :=cmd.Start()	
	if err !=nil{
		fmt.Printf("gst-launch udp error:%s\n",err)
		}
		
	addrRtmp, err2 := net.ResolveUDPAddr("udp", "localhost:"+port)
	if err2 != nil {
		fmt.Printf("net.ResolveUDPAddr %s ", err2)
	}
	conn, err3 := net.DialUDP("udp", nil, addrRtmp)
	if err3 != nil {
		fmt.Printf("net.DialUDP %s ", err3)
	}
	return &RtmpUdp{
		conn:conn,
    }
}

func (r *RtmpUdp)  GetConn()  *net.UDPConn{
	return r.conn
	} 
