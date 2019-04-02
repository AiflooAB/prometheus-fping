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
