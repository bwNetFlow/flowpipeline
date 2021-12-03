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
	"strings"
	"sync"

	"github.com/bwNetFlow/flowpipeline/segments"
	flow "github.com/bwNetFlow/protobuf/go"
)

type Csv struct {
	segments.BaseSegment
	writer     *csv.Writer
	fieldTypes []string
	fieldNames []string

	FileName string // optional, default is empty which means stdout
	Fields   string // optional comma-separated list of fields to export, default is "", meaning all fields
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

	var heading []string
	if config["fields"] != "" {
		protofields := reflect.TypeOf(flow.FlowMessage{})
		conffields := strings.Split(config["fields"], ",")
		for _, field := range conffields {
			protofield, found := protofields.FieldByName(field)
			if !found {
				log.Printf("[error] Csv: Field specified in 'fields' does not exist.")
				return nil
			}
			newsegment.fieldNames = append(newsegment.fieldNames, field)
			newsegment.fieldTypes = append(newsegment.fieldTypes, protofield.Type.String())
			heading = append(heading, field)
		}
	} else {
		protofields := reflect.TypeOf(flow.FlowMessage{})
		// +-3 skips over protobuf state, sizeCache and unknownFields
		newsegment.fieldNames = make([]string, protofields.NumField()-3)
		newsegment.fieldTypes = make([]string, protofields.NumField()-3)
		for i := 3; i < protofields.NumField(); i++ {
			field := protofields.Field(i)
			newsegment.fieldNames[i-3] = field.Name
			newsegment.fieldTypes[i-3] = field.Type.String()
			heading = append(heading, field.Name)
		}
		newsegment.Fields = config["fields"]
	}

	newsegment.writer = csv.NewWriter(file)
	if err := newsegment.writer.Write(heading); err != nil {
		log.Println("[error] Csv: Failed to write to destination:", err)
		return nil
	}
	newsegment.writer.Flush()

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
		for i, fieldname := range segment.fieldNames {
			protofield := values.FieldByName(fieldname)
			switch segment.fieldTypes[i] {
			case "[]uint8": // this is neccessary for proper formatting
				ipstring := net.IP(protofield.Interface().([]uint8)).String()
				if ipstring == "<nil>" {
					ipstring = ""
				}
				record = append(record, ipstring)
			case "uint32": // this is because FormatUint is much faster than Sprint
				record = append(record, strconv.FormatUint(uint64(protofield.Interface().(uint32)), 10))
			case "uint64": // this is because FormatUint is much faster than Sprint
				record = append(record, strconv.FormatUint(uint64(protofield.Interface().(uint64)), 10))
			case "string": // this is because doing nothing is also much faster than Sprint
				record = append(record, protofield.Interface().(string))
			default:
				record = append(record, fmt.Sprint(protofield))
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
