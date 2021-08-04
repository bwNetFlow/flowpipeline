package segments

import (
	"encoding/csv"
	"github.com/bwNetFlow/ip_prefix_trie"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
)

type AddCid struct {
	BaseSegment
	FileName      string
	DropUnmatched bool

	trieV4 ip_prefix_trie.TrieNode
	trieV6 ip_prefix_trie.TrieNode
}

func (segment AddCid) New(config map[string]string) Segment {
	drop, err := strconv.ParseBool(config["dropunmatched"])
	if err != nil {
		log.Println("[warning] AddCid: Config option 'dropunmatched' was not parsable, defaulting to false.")
		drop = false
	}
	return &AddCid{
		FileName:      config["filename"],
		DropUnmatched: drop,
	}
}

func (segment *AddCid) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.out)
		wg.Done()
	}()

	if segment.FileName == "" {
		log.Println("[error] AddCid: This segment requires a 'filename' parameter.")
		os.Exit(1)
	}

	segment.readPrefixList()
	for msg := range segment.in {
		var laddress net.IP
		switch {
		case msg.RemoteAddr == 1: // 1 indicates SrcAddr is the RemoteAddr
			laddress = msg.DstAddr // we want the LocalAddr tho
		case msg.RemoteAddr == 2: // 2 indicates DstAddr is the RemoteAddr
			laddress = msg.SrcAddr // we want the LocalAddr tho
		default:
			if !segment.DropUnmatched {
				segment.out <- msg
			}
			continue
		}

		// prepare matching the address into a prefix and its associated CID
		if laddress.To4() == nil {
			retCid, _ := segment.trieV6.Lookup(laddress).(int64) // try to get a CID
			msg.Cid = uint32(retCid)
		} else {
			retCid, _ := segment.trieV4.Lookup(laddress).(int64) // try to get a CID
			msg.Cid = uint32(retCid)
		}
		segment.out <- msg
	}
}

func (segment *AddCid) readPrefixList() {
	f, err := os.Open(segment.FileName)
	defer f.Close()
	if err != nil {
		log.Printf("[error] AddCid: Could not open prefix list: %v", err)
		return
	}

	csvr := csv.NewReader(f)
	var count int
	for {
		row, err := csvr.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Printf("[warning] AddCid: Encountered non-CSV line in prefix list: %v", err)
				continue
			}
		}

		cid, err := strconv.ParseInt(row[1], 10, 32)
		if err != nil {
			log.Printf("[warning] AddCid: Encountered non-integer customer id: %v", err)
			continue
		}

		// copied from net.IP module to detect v4/v6
		var added bool
		for i := 0; i < len(row[0]); i++ {
			switch row[0][i] {
			case '.':
				segment.trieV4.Insert(cid, []string{row[0]})
				added = true
			case ':':
				segment.trieV6.Insert(cid, []string{row[0]})
				added = true
			}
			if added {
				count += 1
				break
			}
		}
	}
	log.Printf("[info] AddCid: Read prefix list with %d prefixes.", count)
}
