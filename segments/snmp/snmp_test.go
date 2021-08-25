package snmp

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

// SNMPInterface Segment test
// TODO: find a way to run this elsewhere, as this currently only works by
// having the local 161/udp port forwarded to some router.
// func TestSegment_SNMPInterface(t *testing.T) {
// 	result := testSegmentWithFlows(
// 		&SNMPInterface{
// 			Community: "public",
// 			Regex:     "^_[a-z]{3}_[0-9]{5}_[0-9]{5}_ [A-Z0-9]+ (.*?) *( \\(.*)?$",
// 			ConnLimit: 1,
// 		},
// 		[]*flow.FlowMessage{
// 			&flow.FlowMessage{Type: 42, SamplerAddress: []byte{127, 0, 0, 1}, InIf: 70},
// 			&flow.FlowMessage{SamplerAddress: []byte{127, 0, 0, 1}, InIf: 70},
// 		})
// 	if result.SrcIfDesc == "" {
// 		t.Error("Segment SNMPInterface is not adding a SrcIfDesc.")
// 	}
// }
