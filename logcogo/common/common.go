package common

import (
	"net"
	"strings"
)

// 要收集日志的配置项

type CollectConf struct {
	Path  string `json:"path"`
	Topic string `json:"topic"`
}

func GetOutboundIP() (ip string, err error) {
	conn, err := net.Dial("udp", "8.8.8.8:80") //不发送出去 只需要IP
	if err != nil {
		return "", err
	}
	defer conn.Close()

	addr := conn.LocalAddr().(*net.UDPAddr)
	ip = strings.Split(addr.IP.String(), ":")[0]
	return ip, nil
}
