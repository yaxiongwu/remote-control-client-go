1.go build生成可执行文件
2.建一个sh文件,内容如下
    #!/bin/sh
    cd /home/pi/ion-sdk-go-0516/example/ion-sfu-join-from-mediadevice
    nohup ./main &
3.在/etc/rc.local文件的
    exit 0前添加：
    exec 上面的.sh文件
