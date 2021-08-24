package segments

import (
	"log"
	"net"
	"strconv"
	"sync"

	maxmind "github.com/oschwald/maxminddb-golang"
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
		log.Println("[info] GeoLocation: 'dropunmatched' set to default 'false'.")
	}
	if config["filename"] == "" {
		log.Println("[error] GeoLocation: This segment requires the 'filename' parameter.")
		return nil
	}
	newSegment := &GeoLocation{
		FileName:      config["filename"],
		DropUnmatched: drop,
	}
	newSegment.dbHandle, err = maxmind.Open(containerVolumePrefix + config["filename"])
	if err != nil {
		log.Printf("[error] GeoLocation: Could not open specified Maxmind DB file: %v", err)
		return nil
	}
	return newSegment
}

func (segment *GeoLocation) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.out)
		wg.Done()
	}()

	defer func() {
		segment.dbHandle.Close()
	}()

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

func init() {
	segment := &GeoLocation{}
	RegisterSegment("geolocation", segment)
}
