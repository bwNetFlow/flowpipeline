package influx

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	flow "github.com/bwNetFlow/protobuf/go"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"google.golang.org/protobuf/encoding/protojson"
)

// Connector provides export features to Influx
type Connector struct {
	URL          string
	Org          string
	Bucket       string
	Token        string
	ExportFreq   int
	Batchsize    int
	influxClient influxdb2.Client
}

// Initialize a connection to Influxdb
func (c *Connector) Initialize() {
	c.influxClient = influxdb2.NewClientWithOptions(
		c.URL,
		c.Token,
		influxdb2.DefaultOptions().SetBatchSize(uint(c.Batchsize)))

	c.checkBucket()
}

// check if database exists
func (c *Connector) checkBucket() {
	bucket, err := c.influxClient.BucketsAPI().FindBucketByName(context.Background(), c.Bucket)
	if err != nil {
		// TODO: init bucket if not found? Maybe create one?
		log.Printf("[warning] influx: Given bucket %s not found", bucket.Name)
	} else {
		log.Printf("[info] influx: Bucket found with result: %s", bucket.Name)
	}
}

func (c *Connector) CreatePoint(flow *flow.FlowMessage) *write.Point {
	// write tags for datapoint
	// TODO: add more tags e.g. CID
	tags := map[string]string{
		"origin": "belwue",
		"cid":    fmt.Sprint(flow.Cid),
	}

	// marshall protobuf to json
	data, err := protojson.Marshal(flow)
	if err != nil {
		log.Printf("[warning] influx: Skipping a flow, failed to recode protobuf as JSON: %v", err)
		return nil
	}

	// convert json []byte to insert in influx
	fields := make(map[string]interface{})
	err = json.Unmarshal([]byte(data), &fields)
	if err != nil {
		log.Printf("[warning] influx: Skipping a flow, failed to unmarshall JSON: %v", err)
		return nil
	}

	// create point
	p := influxdb2.NewPoint(
		"flowdata",
		tags,
		fields,
		time.Now())
	return p

}
