package pipeline

import (
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/bwNetFlow/flowpipeline/segments"
	"gopkg.in/yaml.v2"
)

// A config representation of a segment. It is intended to look like this:
//   - segment: noop
//     config:
//       key: value
//       foo: bar
// This struct has the appropriate yaml tags inline.
type SegmentRepr struct {
	Name   string            `yaml:"segment"` // to be looked up with a registry
	Config map[string]string `yaml:"config"`  // to be expanded by our instance
}

// Returns the SegmentRepr's Config with all its variables expanded. It tries
// to match numeric variables such as '$1' to the corresponding command line
// argument not matched by flags, or else uses regular environment variable
// expansion.
func (s *SegmentRepr) ExpandedConfig() map[string]string {
	argvMapper := func(placeholderName string) string {
		argnum, err := strconv.Atoi(placeholderName)
		if err == nil && argnum < len(flag.Args()) {
			return flag.Args()[argnum]
		}
		return ""
	}
	expandedConfig := make(map[string]string)
	for k, v := range s.Config {
		expandedConfig[k] = os.Expand(v, argvMapper) // try to convert $n and such to argv[n]
		if expandedConfig[k] == "" && v != "" {      // if unsuccessful, do regular env expansion
			expandedConfig[k] = os.ExpandEnv(v)
		}
	}
	return expandedConfig
}

// Builds a list of Segment objects from raw configuration bytes and
// initializes a Pipeline with them.
func NewFromConfig(config []byte) *Pipeline {
	// parse a list of SegmentReprs from yaml
	pipelineRepr := new([]SegmentRepr)

	err := yaml.Unmarshal(config, &pipelineRepr)
	if err != nil {
		log.Fatalf("[error] Error parsing configuration YAML: %v", err)
	}

	// we have SegmentReprs parsed, instanciate them as actual Segments
	segmentList := make([]segments.Segment, len(*pipelineRepr))
	for i, segmentrepr := range *pipelineRepr {
		segmenttype := segments.LookupSegment(segmentrepr.Name) // a typed nil instance
		if segmenttype == nil {
			os.Exit(1)
		}
		// the Segment's New method knows how to handle our config
		segment := segmenttype.New(segmentrepr.ExpandedConfig())
		if segment != nil {
			segmentList[i] = segment
		} else {
			log.Printf("[error] Configured segment '%s' could not be initialized properly, see previous messages.", segmentrepr.Name)
			os.Exit(1)
		}
	}
	return New(segmentList...)
}
