// Prints all flows to stdout or a given file in json format, for consumption by the stdin segment or for debugging.
package json

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/bwNetFlow/flowpipeline/segments"
	"google.golang.org/protobuf/encoding/protojson"
)

type Json struct {
	segments.BaseSegment
	writer *bufio.Writer

	FileName string // optional, default is empty which means stdout
}

func (segment Json) New(config map[string]string) segments.Segment {
	newsegment := &Json{}

	var filename string = "stdout"
	var file *os.File
	var err error
	if config["filename"] != "" {
		file, err = os.Create(config["filename"])
		if err != nil {
			log.Printf("[error] Json: File specified in 'filename' is not accessible: %s", err)
		}
		filename = config["filename"]
	} else {
		file = os.Stdout
		log.Println("[info] Json: 'filename' unset, using stdout.")
	}
	newsegment.FileName = filename
	newsegment.writer = bufio.NewWriter(file)

	return newsegment
}

func (segment *Json) Run(wg *sync.WaitGroup) {
	defer func() {
		segment.writer.Flush()
		close(segment.Out)
		wg.Done()
	}()
	for msg := range segment.In {
		data, err := protojson.Marshal(msg)
		if err != nil {
			log.Printf("[warning] Json: Skipping a flow, failed to recode protobuf as JSON: %v", err)
			continue
		}

		// use frpintln because it add newline depending on OS
		fmt.Fprintln(segment.writer, string(data))
		// we need to flush here every time because we need full lines and can not wait
		// in case of using this output as in input for other instances consuming flow data
		segment.writer.Flush()
		segment.Out <- msg
	}
}

func init() {
	segment := &Json{}
	segments.RegisterSegment("json", segment)
}
