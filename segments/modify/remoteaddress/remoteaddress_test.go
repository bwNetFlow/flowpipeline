package remoteaddress

import (
	"log"
	"os"
	"testing"

	"github.com/bwNetFlow/flowpipeline/segments"
	flow "github.com/bwNetFlow/protobuf/go"
	"github.com/hashicorp/logutils"
)

func TestMain(m *testing.M) {
	log.SetOutput(&logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"info", "warning", "error"},
		MinLevel: logutils.LogLevel("info"),
		Writer:   os.Stderr,
	})
	code := m.Run()
	os.Exit(code)
}

// RemoteAddress Segment testing is basically checking whether switch/case is working okay...
func TestSegment_RemoteAddress(t *testing.T) {
	result := segments.TestSegment("remoteaddress", map[string]string{"policy": "border"},
		&flow.FlowMessage{FlowDirection: 0})
	if result.RemoteAddr != 1 {
		t.Error("Segment RemoteAddress is not determining RemoteAddr correctly.")
	}
}

func TestSegment_RemoteAddress_localAddrIsDst(t *testing.T) {
	result := segments.TestSegment("remoteaddress", map[string]string{"policy": "cidr", "filename": "../../../examples/enricher/customer_subnets.csv"},
		&flow.FlowMessage{SrcAddr: []byte{192, 168, 88, 42}})
	if result.RemoteAddr != 1 {
		t.Error("Segment RemoteAddress is not determining the local address correctly by 'cidr'.")
	}
}
