// Package csv processes all flows from it's In channel and converts them into
// CSV format. Using it's configuration options it can write to a file or to
// stdout.
package csv

import (
	"encoding/csv"
	"fmt"
	"log"
	"net"
	"os"
	"reflect"
	"strconv"
	"sync"

	"github.com/bwNetFlow/flowpipeline/segments"
	flow "github.com/bwNetFlow/protobuf/go"
)

type Csv struct {
	segments.BaseSegment
	writer     *csv.Writer
	numFields  int
	fieldTypes []string

	FileName string // optional, default is empty which means stdout
}

func (segment Csv) New(config map[string]string) segments.Segment {
	newsegment := &Csv{}

	var filename string = "stdout"
	var file *os.File
	var err error
	if config["filename"] != "" {
		file, err = os.Create(config["filename"])
		if err != nil {
			log.Printf("[error] Csv: File specified in 'filename' is not accessible: %s", err)
		}
		filename = config["filename"]
	} else {
		file = os.Stdout
		log.Println("[info] Csv: 'filename' unset, using stdout.")
	}
	newsegment.FileName = filename

	fields := reflect.TypeOf(flow.FlowMessage{})
	newsegment.numFields = fields.NumField()
	newsegment.fieldTypes = make([]string, newsegment.numFields)
	var heading []string
	for i := 3; i < newsegment.numFields; i++ { // skip over protobuf state, sizeCache and unknownFields
		field := fields.Field(i)
		newsegment.fieldTypes[i] = field.Type.String()
		heading = append(heading, field.Name)
	}

	newsegment.writer = csv.NewWriter(file)
	newsegment.writer.Flush()
	if err := newsegment.writer.Write(heading); err != nil {
		log.Println("[error] Csv: Failed to write to destination:", err)
		return nil
	}

	return newsegment
}

func (segment *Csv) Run(wg *sync.WaitGroup) {
	defer func() {
		segment.writer.Flush()
		close(segment.Out)
		wg.Done()
	}()
	for msg := range segment.In {
		var record []string
		values := reflect.ValueOf(msg).Elem()
		for i := 3; i < segment.numFields; i++ { // skip over protobuf state, sizeCache and unknownFields
			switch segment.fieldTypes[i] {
			case "[]uint8": // this is neccessary for proper formatting
				ipstring := net.IP(values.Field(i).Interface().([]uint8)).String()
				if ipstring == "<nil>" {
					ipstring = ""
				}
				record = append(record, ipstring)
			case "uint32": // this is because FormatUint is much faster than Sprint
				record = append(record, strconv.FormatUint(uint64(values.Field(i).Interface().(uint32)), 10))
			case "uint64": // this is because FormatUint is much faster than Sprint
				record = append(record, strconv.FormatUint(uint64(values.Field(i).Interface().(uint64)), 10))
			case "string": // this is because doing nothing is also much faster than Sprint
				record = append(record, values.Field(i).Interface().(string))
			default:
				record = append(record, fmt.Sprint(values.Field(i)))
			}
		}
		segment.writer.Write(record)
		segment.Out <- msg
	}
}

func init() {
	segment := &Csv{}
	segments.RegisterSegment("csv", segment)
}
