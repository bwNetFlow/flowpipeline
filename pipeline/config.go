package pipeline

import (
	"flag"
	"github.com/bwNetFlow/flowpipeline/segments"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// any new segments need to be added here to be recognized by the config parser
func recognizedInstances() []interface{} {
	return []interface{}{
		segments.AddCid{},
		segments.Count{},
		segments.DropFields{},
		segments.GeoLocation{},
		segments.Goflow{},
		segments.FlowFilter{},
		segments.KafkaConsumer{},
		segments.KafkaProducer{},
		segments.NoOp{},
		segments.PrintFlowdump{},
		segments.PrintDots{},
		segments.StdIn{},
		segments.StdOut{},
	}
}

type SegmentRegistry struct {
	Types map[string]reflect.Type
}

func (s *SegmentRegistry) Lookup(name string) segments.Segment {
	if segmenttype := s.Types[strings.ToLower(name)]; segmenttype != nil {
		v := reflect.New(segmenttype).Elem()
		return v.Addr().Interface().(segments.Segment)
	} else {
		log.Printf("Error: Unknown segment '%s' configured. Options are %+v", name, reflect.ValueOf(s.Types).MapKeys())
		return nil
	}
}

func NewSegmentRegistry() *SegmentRegistry {
	var typemap = make(map[string]reflect.Type)
	for _, v := range recognizedInstances() {
		t := reflect.TypeOf(v)
		// we use lowercase internally and leave the 'segment.' prefix off
		typemap[strings.ToLower(t.String()[9:])] = t
	}
	return &SegmentRegistry{Types: typemap}
}

type SegmentRepr struct {
	Name   string            `yaml:"segment"` // to be looked up with a registry
	Config map[string]string `yaml:"config"`  // to be expanded by our instance
}

func (s *SegmentRepr) ExpandedConfig() (expandedConfig map[string]string) {
	argvMapper := func(placeholderName string) string {
		argnum, err := strconv.Atoi(placeholderName)
		if err == nil && argnum < len(flag.Args()) {
			return flag.Args()[argnum]
		}
		return ""
	}
	expandedConfig = make(map[string]string)
	for k, v := range s.Config {
		expandedConfig[k] = os.Expand(v, argvMapper) // try to convert $n and such to argv[n]
		if expandedConfig[k] == "" && v != "" {      // if unsuccessful, do regular env expansion
			expandedConfig[k] = os.ExpandEnv(v)
		}
	}
	return
}

func SegmentListFromConfig(config []byte) (segmentList []segments.Segment) {
	// parse a list of SegmentReprs from yaml
	pipelineRepr := new([]SegmentRepr)

	err := yaml.Unmarshal(config, &pipelineRepr)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// we have SegmentReprs parsed, instanciate them as actual Segments
	registry := NewSegmentRegistry()
	segmentList = make([]segments.Segment, len(*pipelineRepr))
	for i, segmentrepr := range *pipelineRepr {
		segmenttype := registry.Lookup(segmentrepr.Name) // a typed nil instance
		if segmenttype == nil {
			os.Exit(1)
		}
		// the Segments New method knows how to handle our config
		segmentList[i] = segmenttype.New(segmentrepr.ExpandedConfig())
	}
	return
}

func NewFromConfig(config []byte) *Pipeline {
	return NewPipeline(SegmentListFromConfig(config)...)
}
