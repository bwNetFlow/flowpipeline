package addrstrings

import (
	"github.com/bwNetFlow/flowpipeline/pb"
	"github.com/bwNetFlow/flowpipeline/segments"
	"reflect"
	"testing"
)

func init() {
}

func TestAddrStrings(t *testing.T) {
	tests := map[string]struct {
		input    *pb.EnrichedFlow
		expected *pb.EnrichedFlow
	}{
		"no addresses": {
			input:    &pb.EnrichedFlow{},
			expected: &pb.EnrichedFlow{},
		},
		"full set (IPv4)": {
			input: &pb.EnrichedFlow{
				SrcAddr:        []byte{0x7f, 0x0, 0x0, 0x1},
				DstAddr:        []byte{0x7f, 0x0, 0x0, 0x2},
				NextHop:        []byte{0x7f, 0x0, 0x0, 0x3},
				SamplerAddress: []byte{0x7f, 0x0, 0x0, 0x4},
			},
			expected: &pb.EnrichedFlow{
				SrcAddr:        []byte{0x7f, 0x0, 0x0, 0x1},
				DstAddr:        []byte{0x7f, 0x0, 0x0, 0x2},
				NextHop:        []byte{0x7f, 0x0, 0x0, 0x3},
				SamplerAddress: []byte{0x7f, 0x0, 0x0, 0x4},
				SourceIP:       "127.0.0.1",
				DestinationIP:  "127.0.0.2",
				NextHopIP:      "127.0.0.3",
				SamplerIP:      "127.0.0.4",
			},
		},
		"full set (IPv6)": {
			input: &pb.EnrichedFlow{
				SrcAddr:        []byte{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
				DstAddr:        []byte{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
				NextHop:        []byte{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03},
				SamplerAddress: []byte{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04},
			},
			expected: &pb.EnrichedFlow{
				SrcAddr:        []byte{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
				DstAddr:        []byte{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
				NextHop:        []byte{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03},
				SamplerAddress: []byte{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04},
				SourceIP:       "2001:db8::1",
				DestinationIP:  "2001:db8::2",
				NextHopIP:      "2001:db8::3",
				SamplerIP:      "2001:db8::4",
			},
		},
		"MAC addresses": {
			input: &pb.EnrichedFlow{
				SrcMac: 0x00005e005301,
				DstMac: 0x00005e005302,
			},
			expected: &pb.EnrichedFlow{
				SrcMac:         0x00005e005301,
				DstMac:         0x00005e005302,
				SourceMAC:      "01:53:00:5e:00:00",
				DestinationMAC: "02:53:00:5e:00:00",
			},
		},
	}

	for testname, test := range tests {
		t.Run(testname, func(t *testing.T) {
			result := segments.TestSegment("addrstrings", nil, test.input)
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("Segment addrstrings is not returning the proper fields. Got: »%+v« Expected »%+v«", result, test.expected)
			}
		})
	}
}