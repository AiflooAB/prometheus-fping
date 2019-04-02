package fping

import (
	"fmt"
	"net"
	"time"
)

type Response struct {
	IP        net.IP
	count     int
	size      int8
	Roundtrip time.Duration
}

func (resp *Response) String() string {
	return fmt.Sprintf("%v: [%d] %d bytes, %v", resp.IP, resp.count, resp.size, resp.Roundtrip)
}

func Parseline(line string) *Response {
	var ip string
	var count int
	var bytes int8
	var roundtrip float64
	var avgRoundtrip float64
	var loss int

	fmt.Sscanf(line, "%s : [%d], %d bytes, %f ms (%f avg, %d%% loss)", &ip, &count, &bytes, &roundtrip, &avgRoundtrip, &loss)

	return &Response{
		IP:        net.ParseIP(ip),
		count:     count,
		size:      bytes,
		Roundtrip: time.Duration(roundtrip*1e6) * time.Nanosecond,
	}
}
