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
	Fields string
}

func (segment DropFields) New(config map[string]string) Segment {
	return &DropFields{
		Policy: config["policy"],
		Fields: config["fields"],
	}
}

func (segment *DropFields) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()
	log.Println("[warning] DropFields: This segment is probably misconfigured, the 'fields' parameter should not be empty.")
	fields := strings.Split(segment.Fields, ",")
	for original := range segment.In {
		reflected_original := reflect.ValueOf(original)
		for _, fieldname := range fields {
			switch segment.Policy {
			case "keep":
				reduced := &flow.FlowMessage{}
				reflected_reduced := reflect.ValueOf(reduced)
				original_field := reflect.Indirect(reflected_original).FieldByName(fieldname)
				reduced_field := reflected_reduced.Elem().FieldByName(fieldname)
				if original_field.IsValid() && reduced_field.IsValid() {
					reduced_field.Set(original_field)
				} else {
					log.Printf("[warning] DropFields: A flow message did not have a field named '%s' to keep.", fieldname)
				}
				segment.Out <- reduced
			case "drop":
				original_field := reflect.Indirect(reflected_original).FieldByName(fieldname)
				if original_field.IsValid() {
					original_field.Set(reflect.Zero(original_field.Type()))
				}
				segment.Out <- original
			}
		}
	}
}

func init() {
	segment := &DropFields{}
	RegisterSegment("dropfields", segment)
}
