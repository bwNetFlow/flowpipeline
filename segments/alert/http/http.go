// This segment is used to alert on flows using webhooks - WIP, but basically usable.
package http

import (
	"bytes"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/bwNetFlow/flowpipeline/segments"
	"google.golang.org/protobuf/encoding/protojson"
)

type Http struct {
	segments.BaseSegment
	Url string
	// TODO: add async parameter
	// TODO: add timeout options
	// TODO: add more HTTP methods
	// TODO: add ability to include custom message
	// TODO: add conditional http send?
	// TODO: add ability to limit data sent?
}

func (segment Http) New(config map[string]string) segments.Segment {
	requestUrl, err := url.Parse(config["url"])
	if err != nil {
		log.Printf("[error] Http: error parsing url parameter: %e", err)
		return nil
	}
	if !(requestUrl.Scheme == "http" || requestUrl.Scheme == "https") {
		log.Printf("[error] Http: error parsing url parameter, scheme must be 'http://' or 'https://'")
		return nil
	}
	return &Http{Url: config["url"]}
}

func (segment *Http) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()
	var limitLog bool
	for msg := range segment.In {
		data, err := protojson.Marshal(msg)
		if err != nil {
			log.Printf("[warning] Http: Skipping a flow, failed to recode protobuf as JSON: %v", err)
			continue
		}

		resp, err := http.Post(segment.Url, "application/json", bytes.NewBuffer(data))
		if err != nil {
			log.Printf("[error] Http: Request setup error, skipping at least one flow: %v", err)
			log.Print("[error] Above message will not repeat for every flow and is effective until resolved.")
			limitLog = true
		} else if !(resp.StatusCode-200 < 100) {
			log.Printf("[error] Http: Server endpoint error, skipping at least one flow. Code %s.", resp.Status)
			log.Print("[error] Above message will not repeat for every flow and is effective until resolved.")
			limitLog = true
		} else if limitLog {
			log.Print("[resolved] Http: Previous error is resolved, flows are being posted to configured url successfully again.")
			limitLog = false
		}
		segment.Out <- msg
	}
}

func init() {
	segment := &Http{}
	segments.RegisterSegment("http", segment)
}
