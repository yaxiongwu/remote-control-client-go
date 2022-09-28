1.go build生成可执行文件
2.建一个sh文件,内容如下
    #!/bin/sh
    cd /home/pi/目标地址
    nohup ./main &
3.chmod +x *.sh
  chmod +x 需要执行的文件main
  chmod +x /etc/rc.local 
4.在/etc/rc.local文件的
    exit 0前添加：
    sleep 10
    su pi -c "bash /home/pi/remote-control-client-go-0927/rpi/autoRunWebrtc.sh"
