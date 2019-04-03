package fping

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

type Response struct {
	IP        net.IP
	count     int
	size      int8
	Roundtrip time.Duration
}

type UnreachableResponse struct {
	IP net.IP
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

func ParseStderr(line string) *UnreachableResponse {
	if len(line) == 0 {
		return nil
	}
	if strings.Contains(line, "ICMP Host Unreachable") {
		var from string
		var to string
		fmt.Sscanf(line, "ICMP Host Unreachable from %s for ICMP Echo sent to %s", &from, &to)
		return &UnreachableResponse{
			IP: net.ParseIP(to),
		}
	}
	// Ignore summary lines
	if strings.Contains(line, "xmt/rcv/%loss") {
		return nil
	}
	fmt.Fprintf(os.Stderr, "Failed to parse line: '%s'\n", line)
	return nil
}
