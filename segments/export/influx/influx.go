package influx

import (
	"sync"

	"github.com/bwNetFlow/flowpipeline/segments"
)

type Influx struct {
	segments.BaseSegment
	URL    string // URL to influxdb endpoint
	Org    string // influx org name
	Bucket string // influx bucket
	Token  string
	// 	Batchsize  uint32 // set the batch size for the writer
	// 	ExportFreq uint32 // set the frequency for the writer
}

func (segment Influx) New(config map[string]string) segments.Segment {
	// TODO: add paramteres for Influx endpoint and eval vars
	var url = "127.0.0.1:8086"
	if config["url"] != "" {
		url = config["url"]
	}
	var org = "my-org"
	if config["org"] != "" {
		org = config["org"]
	}
	var bucket = "my-bucket"
	if config["bucket"] != "" {
		bucket = config["bucket"]
	}
	var token = "my-token"
	if config["token"] != "" {
		token = config["token"]
	}

	return &Influx{
		URL:    url,
		Org:    org,
		Bucket: bucket,
		Token:  token,
	}
}

func (segment *Influx) Run(wg *sync.WaitGroup) {
	// TODO: extend options
	var connector = Connector{URL: segment.URL, Bucket: segment.Bucket, Org: segment.Org, Token: segment.Token, Batchsize: 5000}
	// initialize Influx endpoint
	connector.Initialize()
	writeAPI := connector.influxClient.WriteAPI(connector.Org, connector.Bucket)
	defer func() {
		close(segment.Out)
		// Force all unwritten data to be sent
		writeAPI.Flush()
		connector.influxClient.Close()
		wg.Done()
	}()

	for msg := range segment.In {
		segment.Out <- msg
		datapoint := connector.CreatePoint(msg)
		if datapoint == nil {
			// just ignore raised warnings if flow cannot be converted or unmarshalled
			continue
		}
		// async write
		writeAPI.WritePoint(datapoint)
	}
}

func init() {
	segment := &Influx{}
	segments.RegisterSegment("influx", segment)
}
