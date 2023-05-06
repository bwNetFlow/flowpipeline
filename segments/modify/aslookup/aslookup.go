package aslookup

import (
    "os"
	"log"
    "net"
	"sync"

    "github.com/banviktor/asnlookup/pkg/database"
	"github.com/bwNetFlow/flowpipeline/segments"
)

type AsLookup struct {
	segments.BaseSegment
    FileName    string

    asDatabase  database.Database
}

func (segment AsLookup) New(config map[string]string) segments.Segment {

    // open lookup file
	lookupfile, err := os.OpenFile(config["filename"], os.O_RDONLY, 0)
	if err != nil {
		log.Printf("[error] AsLookup: Error opening lookup file: %s", err)
		return nil
	}
    defer lookupfile.Close()

    newSegment := &AsLookup{
        FileName: config["filename"],
    }

	// lookup file can either be an MRT file or a lookup database generated with asnlookup
    // see: https://github.com/banviktor/asnlookup
    if config["type"] == "db" {
        // open lookup db
        db, err := database.NewFromDump(lookupfile)
        if err != nil {
            log.Printf("[error] AsLookup: Error parsing database file: %s", err)
        }
        newSegment.asDatabase = db
    } else {
        // parse with asnlookup
        builder := database.NewBuilder()
        if err = builder.ImportMRT(lookupfile); err != nil {
            log.Printf("[error] AsLookup: Error parsing MRT file: %s", err)
        }

        // build lookup database
        db, err := builder.Build()
        if err != nil {
            log.Printf("[error] AsLookup: Error building lookup database: %s", err)
        }
        newSegment.asDatabase = db
    }


	return newSegment
}

func (segment *AsLookup) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()
	for msg := range segment.In {
		// Look up destination AS
        dstIp := net.ParseIP(msg.DstAddrObj().String())
        dstAs, err := segment.asDatabase.Lookup(dstIp)
        if err != nil {
            log.Printf("[error] AsLookup: Error looking up ASN for %s: %s", msg.DstAddrObj().String(), err)
		    segment.Out <- msg
            continue
        }
        msg.DstAS = dstAs.Number

		// Look up source AS
        srcIp := net.ParseIP(msg.SrcAddrObj().String())
        srcAs, err := segment.asDatabase.Lookup(srcIp)
        if err != nil {
            log.Printf("[error] AsLookup: Error looking up ASN for %s: %s", msg.SrcAddrObj().String(), err)
		    segment.Out <- msg
            continue
        }
        msg.SrcAS = srcAs.Number

		segment.Out <- msg
	}
}

func init() {
	segment := &AsLookup{}
	segments.RegisterSegment("aslookup", segment)
}
