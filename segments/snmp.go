package segments

import (
	"log"
	"net"
	"regexp"
	"sync"
	"time"

	cache "github.com/patrickmn/go-cache"
)

type SNMPInterface struct {
	BaseSegment
	Community     string
	Regex         string
	compiledRegex *regexp.Regexp
	snmpCache     *cache.Cache
}

func (segment SNMPInterface) New(config map[string]string) Segment {
	return &SNMPInterface{
		Community: config["community"],
		Regex:     config["regex"],
	}
}

func (segment *SNMPInterface) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.out)
		wg.Done()
	}()

	// init cache:			expiry       purge
	segment.snmpCache = cache.New(1*time.Hour, 1*time.Hour)

	var err error
	segment.compiledRegex, err = regexp.Compile(segment.Regex)
	if err != nil {
		log.Printf("[error] SNMPInterface: Configuration error, regex does not compile: %v", err)
	}

	for msg := range segment.in {
		router := net.IP(msg.SamplerAddress).String()
		if msg.InIf > 0 {
			msg.SrcIfName, msg.SrcIfDesc, msg.SrcIfSpeed = fetchInterfaceData(router, msg.InIf)
		}
		if msg.OutIf > 0 {
			msg.DstIfName, msg.DstIfDesc, msg.DstIfSpeed = fetchInterfaceData(router, msg.OutIf)
		}
		segment.out <- msg
	}
}

// Query a single SNMP datapoint. Supposedly a short-lived goroutine.
func querySNMP(router string, iface uint32, datapoint string) {
}

func fetchInterfaceData(router string, iface uint32) (string, string, uint32) {
	return "", "", 0
}

func init() {
	segment := &SNMPInterface{}
	RegisterSegment("snmpinterface", segment)
}
