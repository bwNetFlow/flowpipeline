package segments

import (
	maxmind "github.com/oschwald/maxminddb-golang"
	"log"
	"net"
	"strconv"
	"sync"
)

type GeoLocation struct {
	BaseSegment
	FileName      string
	DropUnmatched bool

	dbHandle *maxmind.Reader
}

func (segment GeoLocation) New(config map[string]string) Segment {
	drop, err := strconv.ParseBool(config["dropunmatched"])
	if err != nil {
		log.Println("[warning] GeoLocation: Config option 'dropunmatched' was not parsable, defaulting to false.")
		drop = false
	}
	return &GeoLocation{
		FileName:      config["filename"],
		DropUnmatched: drop,
	}
}

func (segment *GeoLocation) Run(wg *sync.WaitGroup) {
	defer func() {
		segment.dbHandle.Close()
		close(segment.out)
		wg.Done()
	}()

	segment.readDb()

	var dbrecord struct {
		Country struct {
			ISOCode string `maxminddb:"iso_code"`
		} `maxminddb:"country"`
	}

	for msg := range segment.in {
		var raddress net.IP
		switch {
		case msg.RemoteAddr == 1: // 1 indicates SrcAddr is the RemoteAddr
			raddress = msg.SrcAddr
		case msg.RemoteAddr == 2: // 2 indicates DstAddr is the RemoteAddr
			raddress = msg.DstAddr
		default:
			if !segment.DropUnmatched {
				segment.out <- msg
			}
			continue
		}

		err := segment.dbHandle.Lookup(raddress, &dbrecord)
		if err == nil {
			msg.RemoteCountry = dbrecord.Country.ISOCode
		} else {
			log.Printf("[error] GeoLocation: Lookup of remote address failed: %v", err)
		}
		segment.out <- msg
	}
}

func (segment *GeoLocation) readDb() {
	var err error
	segment.dbHandle, err = maxmind.Open(segment.FileName)
	if err != nil {
		log.Printf("[error] GeoLocation: Could not open specified Maxmind DB file: %v", err)
		return
	}
}
