package geolocation

import (
	"io/ioutil"
	"log"
	"os"
	"sync"
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

func TestSegment_GeoLocation_both(t *testing.T) {
	result := segments.TestSegment("geolocation", map[string]string{"filename": "../../../examples/enricher/GeoLite2-Country-Test.mmdb", "matchboth": "1"},
		&flow.FlowMessage{DstAddr: []byte{2, 125, 160, 218}})
	if result == nil || result.DstCountry != "GB" {
		t.Error("Segment GeoLocation is not adding DstCountry correctly.")
	}
}

// GeoLocation Segment benchmark passthrough
func BenchmarkGeoLocation(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	os.Stdout, _ = os.Open(os.DevNull)

	segment := GeoLocation{}.New(map[string]string{"filename": "../../../examples/enricher/GeoLite2-Country-Test.mmdb"})

	in, out := make(chan *flow.FlowMessage), make(chan *flow.FlowMessage)
	segment.Rewire([]chan *flow.FlowMessage{in, out}, 0, 1)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go segment.Run(wg)

	for n := 0; n < b.N; n++ {
		in <- &flow.FlowMessage{RemoteAddr: 2, DstAddr: []byte{2, 125, 160, 218}}
		_ = <-out
	}
	close(in)
}
