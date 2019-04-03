package fping

import (
	"net"
	"reflect"
	"testing"
	"time"
)

func TestParseLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected *Response
	}{
		{
			"Basic parsing",
			"192.168.1.1   : [12], 84 bytes, 1.29 ms (1.29 avg, 0% loss)",
			&Response{
				IP:        net.ParseIP("192.168.1.1"),
				count:     12,
				size:      84,
				Roundtrip: 1290 * time.Microsecond,
			}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed := Parseline(test.line)
			if !reflect.DeepEqual(test.expected, parsed) {
				t.Errorf("%+v != %+v", test.expected, parsed)
			}
		})
	}
}

func TestParseStderr(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected *UnreachableResponse
	}{
		{
			"Basic parsing",
			"ICMP Host Unreachable from 192.168.9.28 for ICMP Echo sent to 192.168.9.203",
			&UnreachableResponse{IP: net.ParseIP("192.168.9.203")},
		},
		{
			"Ignore empty line",
			"",
			nil,
		},
		{
			"Ignore summary lines",
			"192.168.8.11  : xmt/rcv/%loss = 1/1/0%, min/avg/max = 2.09/2.09",
			nil,
		},
		{
			"Ignore random line",
			"I have no idea what the contents here should be, but we should ignore it",
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed := ParseStderr(test.line)
			if !reflect.DeepEqual(test.expected, parsed) {
				t.Errorf("%+v != %+v", test.expected, parsed)
			}
		})
	}
}
