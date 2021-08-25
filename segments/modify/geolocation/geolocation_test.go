package geolocation

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

// GeoLocation Segment tests are thorough and try every combination
func TestSegment_GeoLocation_noRemoteAddrKeep(t *testing.T) {
	result := segments.TestSegment("geolocation", map[string]string{"filename": "../../../examples/enricher/GeoLite2-Country-Test.mmdb"},
		&flow.FlowMessage{RemoteAddr: 0, SrcAddr: []byte{2, 125, 160, 218}, DstAddr: []byte{2, 125, 160, 218}})
	if result.RemoteCountry != "" {
		t.Error("Segment GeoLocation is adding a RemoteCountry when the remote address is undetermined.")
	}
}

func TestSegment_GeoLocation_noRemoteAddrDrop(t *testing.T) {
	result := segments.TestSegment("geolocation", map[string]string{"filename": "../../../examples/enricher/GeoLite2-Country-Test.mmdb", "dropunmatched": "1"},
		&flow.FlowMessage{RemoteAddr: 0, SrcAddr: []byte{2, 125, 160, 218}, DstAddr: []byte{2, 125, 160, 218}})
	if result != nil {
		t.Error("Segment GeoLocation is not dropping the flow as instructed if the remote address is undetermined.")
	}
}

func TestSegment_GeoLocation_remoteAddrIsSrc(t *testing.T) {
	result := segments.TestSegment("geolocation", map[string]string{"filename": "../../../examples/enricher/GeoLite2-Country-Test.mmdb"},
		&flow.FlowMessage{RemoteAddr: 1, SrcAddr: []byte{2, 125, 160, 218}})
	if result.RemoteCountry != "GB" {
		t.Error("Segment GeoLocation is not adding RemoteCountry when the remote address is the source address.")
	}
}

func TestSegment_GeoLocation_remoteAddrIsDst(t *testing.T) {
	result := segments.TestSegment("geolocation", map[string]string{"filename": "../../../examples/enricher/GeoLite2-Country-Test.mmdb"},
		&flow.FlowMessage{RemoteAddr: 2, DstAddr: []byte{2, 125, 160, 218}})
	if result == nil || result.RemoteCountry != "GB" {
		t.Error("Segment GeoLocation is not adding RemoteCountry when the remote address is the destination address.")
	}
}
