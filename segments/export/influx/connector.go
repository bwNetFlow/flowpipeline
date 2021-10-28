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
)

// Connector provides export features to Influx
type Connector struct {
	Address      string
	Org          string
	Bucket       string
	Token        string
	Tag          string
	ExportFreq   int
	Batchsize    int
	influxClient influxdb2.Client
}

// Initialize a connection to Influxdb
func (c *Connector) Initialize() {
	c.influxClient = influxdb2.NewClientWithOptions(
		c.Address,
		c.Token,
		influxdb2.DefaultOptions().SetBatchSize(uint(c.Batchsize)))

	c.checkBucket()
}

// check if database exists
func (c *Connector) checkBucket() {
	bucket, err := c.influxClient.BucketsAPI().FindBucketByName(context.Background(), c.Bucket)
	if err != nil {
		// The bucket should be created by the Influxdb admin.
		log.Printf("[warning] influx: Given bucket %s not found.", c.Bucket)
	} else {
		log.Printf("[info] influx: Bucket found with result: %s", bucket.Name)
	}
}

func (c *Connector) CreatePoint(flow *flow.FlowMessage) *write.Point {
	// write tags for datapoint
	// TODO: make tags configurable
	tags := map[string]string{
		"origin": c.Tag,
		"cid":    fmt.Sprint(flow.Cid),
	}

	// marshall protobuf to json
	data, err := json.Marshal(flow)
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
