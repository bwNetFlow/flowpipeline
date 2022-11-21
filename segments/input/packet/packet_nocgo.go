//go:build !cgo
// +build !cgo

package packet

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

const cgoEnabled = false
const pfringEnabled = false

type dummyHandle interface {
	LinkType() layers.LinkType
	Close()
	ReadPacketData() (data []byte, ci gopacket.CaptureInfo, err error)
}

var getPcapHandle func(source string, filter string) dummyHandle
var getPcapFile func(source string, filter string) dummyHandle
var getPfringHandle func(source string, filter string) dummyHandle
