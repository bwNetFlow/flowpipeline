//go:build cgo && pfring
// +build cgo,pfring

package packet

const cgoEnabled = true
const pfringEnabled = true

func getPfringHandle(source string, filter string) *pfring.Ring {
	var ring *pfring.Ring
	var err error
	if ring, err = pfring.NewRing(source, 65536, pfring.FlagPromisc); err != nil {
		log.Fatalf("[error]: Packet: Could not setup pfring capture: %v", err)
	} else if err := ring.SetBPFFilter(filter); err != nil {
		log.Fatalf("[error]: Packet: Could not set BPF filter: %v", err)
	} else if err := ring.Enable(); err != nil {
		log.Fatalf("[error]: Packet: Could not initiate capture: %v", err)
	}
	return ring
}
