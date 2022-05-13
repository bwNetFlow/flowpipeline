package main

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/asecurityteam/rolling"
	"github.com/bwNetFlow/flowpipeline/segments"
	"github.com/dustin/go-humanize"
)

type Record struct {
	DstIp       string
	LastUpdated time.Time
	Bytes       *rolling.TimePolicy
	Packets     *rolling.TimePolicy
}

type Ddos struct {
	segments.BaseSegment
}

func (segment Ddos) New(config map[string]string) segments.Segment {
	return &Ddos{}
}

func (segment *Ddos) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()
	database := map[string]*Record{}

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ticker.C:
			databaseEntries := []*Record{}
			for key, entry := range database {
				if entry.LastUpdated.Before(time.Now().Add(-5 * time.Minute)) {
					delete(database, key)
				}
				databaseEntries = append(databaseEntries, entry)
			}
			sort.Slice(databaseEntries, func(i, j int) bool {
				iBytes := databaseEntries[i].Bytes.Reduce(rolling.Sum)
				jBytes := databaseEntries[j].Bytes.Reduce(rolling.Sum)
				return iBytes > jBytes
			})
			printedRecords := 0
			for _, record := range databaseEntries {
				fmt.Printf("%s: %s, %s\n",
					record.DstIp,
					humanize.SI(record.Bytes.Reduce(rolling.Sum)*8/60, "bps"),
					humanize.SI(record.Packets.Reduce(rolling.Sum)*8/60, "pps"),
				)
				printedRecords += 1
				if printedRecords >= 25 {
					break
				}
			}
			fmt.Println("===================================================================")
		case msg, ok := <-segment.In:
			if !ok {
				return
			}
			record := database[msg.DstAddrObj().String()]
			if record == nil {
				record = &Record{Bytes: rolling.NewTimePolicy(rolling.NewWindow(60), time.Second), Packets: rolling.NewTimePolicy(rolling.NewWindow(60), time.Second)}
				record.DstIp = msg.DstAddrObj().String()
			}
			record.Bytes.Append(float64(msg.Bytes))
			record.Packets.Append(float64(msg.Packets))
			record.LastUpdated = time.Now()
			database[msg.DstAddrObj().String()] = record

			segment.Out <- msg
		}
	}
}

func init() {
	segment := &Ddos{}
	segments.RegisterSegment("ddos", segment)
}
