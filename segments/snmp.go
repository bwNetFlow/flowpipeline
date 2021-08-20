package segments

import (
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/alouca/gosnmp"
	cache "github.com/patrickmn/go-cache"
)

var (
	oidBase = ".1.3.6.1.2.1.31.1.1.1.%d.%d"
	oidExts = map[string]uint8{"name": 1, "speed": 15, "desc": 18}
)

type SNMPInterface struct {
	BaseSegment
	Community string
	Regex     string
	ConnLimit uint64

	compiledRegex      *regexp.Regexp
	snmpCache          *cache.Cache
	connLimitSemaphore chan struct{}
}

func (segment SNMPInterface) New(config map[string]string) Segment {
	var connLimit uint64 = 16
	if parsedConnLimit, err := strconv.ParseUint(config["connlimit"], 10, 32); err == nil {
		if parsedConnLimit == 0 {
			log.Println("[error] SNMPInterface: Limiting connections to 0 will not work. Remove this segment or use a higher value (recommendation >= 16).")
		}
		connLimit = parsedConnLimit
	}
	return &SNMPInterface{
		Community: config["community"],
		Regex:     config["regex"],
		ConnLimit: connLimit,
	}
}

func (segment *SNMPInterface) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.out)
		wg.Done()
	}()

	// init cache:			expiry       purge
	segment.snmpCache = cache.New(1*time.Hour, 1*time.Hour) // TODO: make configurable
	// init semaphore for connection limit
	segment.connLimitSemaphore = make(chan struct{}, segment.ConnLimit)

	var err error
	segment.compiledRegex, err = regexp.Compile(segment.Regex)
	if err != nil {
		log.Printf("[error] SNMPInterface: Configuration error, regex does not compile: %v", err)
	}

	for msg := range segment.in {
		router := net.IP(msg.SamplerAddress).String()
		// TODO: rename SrcIf and DstIf fields to match goflow InIf/OutIf
		if msg.InIf > 0 {
			msg.SrcIfName, msg.SrcIfDesc, msg.SrcIfSpeed = segment.fetchInterfaceData(router, msg.InIf)
			if msg.SrcIfDesc != "" {
				cleanDesc := segment.compiledRegex.FindStringSubmatch(msg.SrcIfDesc)
				if cleanDesc != nil && len(cleanDesc) > 1 {
					msg.SrcIfDesc = cleanDesc[1]
				}
			}
		}
		if msg.OutIf > 0 {
			msg.DstIfName, msg.DstIfDesc, msg.DstIfSpeed = segment.fetchInterfaceData(router, msg.OutIf)
			if msg.DstIfDesc != "" {
				cleanDesc := segment.compiledRegex.FindStringSubmatch(msg.DstIfDesc)
				if cleanDesc != nil && len(cleanDesc) > 1 {
					msg.DstIfDesc = cleanDesc[1]
				}
			}
		}
		segment.out <- msg
	}
}

// Query a single SNMP datapoint. Supposedly a short-lived goroutine.
func (segment *SNMPInterface) querySNMP(router string, iface uint32, key string) {
	defer func() {
		<-segment.connLimitSemaphore // release
	}()
	segment.connLimitSemaphore <- struct{}{} // acquire

	s, err := gosnmp.NewGoSNMP(router, segment.Community, gosnmp.Version2c, 1)
	if err != nil {
		log.Println("[error] SNMPInterface: Connection Error:", err)
		segment.snmpCache.Delete(fmt.Sprintf("%s-%d-%s", router, iface, key))
		return
	}

	var result *gosnmp.SnmpPacket
	oid := fmt.Sprintf(oidBase, oidExts[key], iface)
	resp, err := s.Get(oid)
	if err != nil {
		log.Printf("[warning] SNMPInterface: Failed getting %s from %s. Error: %s", key, router, err)
		segment.snmpCache.Delete(fmt.Sprintf("%s-%d-%s", router, iface, key))
		return
	} else {
		result = resp
	}

	// parse and cache
	if len(result.Variables) == 1 {
		snmpvalue := resp.Variables[0].Value
		segment.snmpCache.Set(fmt.Sprintf("%s-%d-%s", router, iface, key), snmpvalue, cache.DefaultExpiration)
	} else {
		log.Printf("[warning] SNMPInterface: Bad response getting %s from %s. Error: %v", key, router, resp.Variables)
	}
}

// Fetch interface data from cache or from the live router. The latter is done
// async, so this method will return nils on the first call for any specific interface.
func (segment *SNMPInterface) fetchInterfaceData(router string, iface uint32) (string, string, uint32) {
	var name, desc string
	var speed uint32
	for key, _ := range oidExts {
		// if value in cache and cache content is not nil, i.e. marked as "being queried"
		if value, found := segment.snmpCache.Get(fmt.Sprintf("%s-%d-%s", router, iface, key)); found && value != nil {
			switch key {
			case "name":
				name = value.(string)
			case "desc":
				desc = value.(string)
			case "speed":
				speed = uint32(value.(uint64))
			}
		} else {
			// mark as "being queried" by putting nil into the cache
			segment.snmpCache.Set(fmt.Sprintf("%s-%d-%s", router, iface, key), nil, cache.DefaultExpiration)
			// go query it
			go segment.querySNMP(router, iface, key)
		}
	}
	return name, desc, speed
}

func init() {
	segment := &SNMPInterface{}
	RegisterSegment("snmpinterface", segment)
}
