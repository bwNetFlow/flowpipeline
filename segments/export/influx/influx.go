// Collects and exports all flows to influxdb for long term storage.
package influx

import (
	"log"
	"net/url"
	"sync"

	"github.com/bwNetFlow/flowpipeline/segments"
)

type Influx struct {
	segments.BaseSegment
	Address string // URL for influxdb endpoint
	Org     string // influx org name
	Bucket  string // influx bucket
	Token   string // influx access token
	Tag     string // optional, tag for Datapoints to insert in influx
	// 	Batchsize  uint32 // set the batch size for the writer
	// 	ExportFreq uint32 // set the frequency for the writer
}

func (segment Influx) New(config map[string]string) segments.Segment {
	// TODO: add paramteres for Influx endpoint and eval vars
	var address = "http://127.0.0.1:8086"
	if config["address"] != "" {
		address = config["address"]
		// check if a valid url has been passed
		_, err := url.Parse(address)
		if err != nil {
			log.Printf("[error] Influx: error parsing given url: %e", err)
		}
		address = config["address"]
	} else {
		log.Println("[info] Influx: Missing configuration parameter 'address'. Using default address 'http://127.0.0.1:8086'")
	}

	var org string
	if config["org"] == "" {
		log.Println("[error] Influx: Missing configuration parameter 'org'. Please set the organization to use.")
	} else {
		org = config["org"]
	}

	var bucket string
	if config["bucket"] == "" {
		log.Println("[error] Influx: Missing configuration parameter 'bucket'. Please set the bucket to use.")
	} else {
		bucket = config["bucket"]
	}

	var token string
	if config["token"] == "" {
		log.Println("[error] Influx: Missing configuration parameter 'token'. Please set the token to use.")
	} else {
		token = config["token"]
	}

	var tag = "my-custom-tag"
	if config["tag"] == "" {
		log.Println("[info] Influx: Missing configuration parameter 'tag'. Using default tag 'origin' with value 'my-custom-tag'")
	} else {
		tag = config["tag"]
		log.Printf("[info] Influx: Using Tag 'origin' with given vaule '%s'", tag)
	}

	return &Influx{
		Address: address,
		Org:     org,
		Bucket:  bucket,
		Token:   token,
		Tag:     tag,
	}
}

func (segment *Influx) Run(wg *sync.WaitGroup) {
	// TODO: extend options
	var connector = Connector{
		Address:   segment.Address,
		Bucket:    segment.Bucket,
		Org:       segment.Org,
		Token:     segment.Token,
		Tag:       segment.Tag,
		Batchsize: 5000,
	}

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
