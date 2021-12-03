package bpf

import (
	"log"
	"os"
	"testing"

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

// Bpf Segment test, passthrough test TODO: how to guarantee device presence on any host
// func TestSegment_Bpf_passthrough(t *testing.T) {
// 	result := segments.TestSegment("bpf", map[string]string{"device": "eth0"},
// 		&flow.FlowMessage{Type: 3})
// 	if result.Type != 3 {
// 		t.Error("Segment Bpf is not working.")
// 	}
// }
