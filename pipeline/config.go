package pipeline

import (
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/bwNetFlow/flowpipeline/segments"
	"gopkg.in/yaml.v2"
)

type SegmentRepr struct {
	Name   string            `yaml:"segment"` // to be looked up with a registry
	Config map[string]string `yaml:"config"`  // to be expanded by our instance
}

func (s *SegmentRepr) ExpandedConfig() map[string]string {
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
	return expandedConfig
}

func SegmentListFromConfig(config []byte) []segments.Segment {
	// parse a list of SegmentReprs from yaml
	pipelineRepr := new([]SegmentRepr)

	err := yaml.Unmarshal(config, &pipelineRepr)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// we have SegmentReprs parsed, instanciate them as actual Segments
	segmentList = make([]segments.Segment, len(*pipelineRepr))
	for i, segmentrepr := range *pipelineRepr {
		segmenttype := segments.LookupSegment(segmentrepr.Name) // a typed nil instance
		if segmenttype == nil {
			os.Exit(1)
		}
		// the Segment's New method knows how to handle our config
		segmentList[i] = segmenttype.New(segmentrepr.ExpandedConfig())
	}
	return segmentList
}

func NewFromConfig(config []byte) *Pipeline {
	return NewPipeline(SegmentListFromConfig(config)...)
}
