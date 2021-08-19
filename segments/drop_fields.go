package segments

import (
	"log"
	"reflect"
	"strings"
	"sync"

	flow "github.com/bwNetFlow/protobuf/go"
)

type DropFields struct {
	BaseSegment
	Policy string
	Keep   string
	Drop   string
}

func (segment DropFields) New(config map[string]string) Segment {
	return &DropFields{
		Policy: config["policy"],
		Keep:   config["keep"],
		Drop:   config["drop"],
	}
}

func (segment *DropFields) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.out)
		wg.Done()
	}()
	if segment.Policy == "keep" && segment.Keep == "" {
		log.Println("[warning] DropFields: This segment is probably misconfigured, if 'policy: keep' and 'keep: \"\"' all flows will be empty.")
	} else if segment.Policy == "drop" && segment.Drop == "" {
		log.Println("[warning] DropFields: This segment is probably misconfigured, if 'policy: drop', the 'drop' parameter needs to be set for the segment to do anything.")
	}
	for original := range segment.in {
		reflected_original := reflect.ValueOf(original)
		switch segment.Policy {
		case "keep":
			reduced := &flow.FlowMessage{}
			reflected_reduced := reflect.ValueOf(reduced)
			for _, fieldname := range strings.Split(segment.Keep, ",") {
				original_field := reflect.Indirect(reflected_original).FieldByName(fieldname)
				reduced_field := reflected_reduced.Elem().FieldByName(fieldname)
				if original_field.IsValid() && reduced_field.IsValid() {
					reduced_field.Set(original_field)
				} else {
					log.Printf("[warning] DropFields: A flow message did not have a field named '%s' to keep.", fieldname)
				}
			}
			segment.out <- reduced
		case "drop":
			for _, fieldname := range strings.Split(segment.Drop, ",") {
				original_field := reflect.Indirect(reflected_original).FieldByName(fieldname)
				if original_field.IsValid() {
					original_field.Set(reflect.Zero(original_field.Type()))
				}
			}
			segment.out <- original
		}
	}
}

func init() {
	segment := &DropFields{}
	RegisterSegment("dropfields", segment)
}
