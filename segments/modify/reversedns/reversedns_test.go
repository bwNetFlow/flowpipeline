package reversedns

import (
	"testing"

	"github.com/bwNetFlow/flowpipeline/pb"
	"github.com/bwNetFlow/flowpipeline/segments"
)

// TODO: create test with a fixes reverse dns address

// ReverseDns Segment test, passthrough
func TestSegment_ReverseDns_passthrough(t *testing.T) {
	result := segments.TestSegment("reversedns", map[string]string{},
		&pb.EnrichedFlow{Bytes: 1})
	if result.Bytes != 1 {
		t.Error("Segment ReverseDns is not working correctly.")
	}
}

func TestSegment_ReverseDns_resolve(t *testing.T) {
	result := segments.TestSegment("reversedns", map[string]string{},
		&pb.EnrichedFlow{Bytes: 1, SrcAddr: []byte{8, 8, 8, 8}})
	if result.Bytes != 1 || result.SrcHostName != "dns.google." {
		t.Errorf("Segment ReverseDns is not resolving correctly. Got %s, expected dns.google", result.SrcHostName)
	}
}
