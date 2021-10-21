package skip

import (
	"log"
	"strconv"
	"sync"

	"github.com/bwNetFlow/flowfilter/parser"
	"github.com/bwNetFlow/flowfilter/visitors"
	"github.com/bwNetFlow/flowpipeline/segments"
	flow "github.com/bwNetFlow/protobuf/go"
)

type Skip struct {
	segments.BaseSegment
	AltOut chan<- *flow.FlowMessage // in addition to the BaseSegment declarations, a second Out is provided

	Condition string // optional, default "", if this is evaluated as true, a skip will occur
	Skip      uint   // optional, default 1, the number of segments to skip after this one when condition is true
	Invert    bool   // optional, default false, whether to invert the conditional behaviour

	expression *parser.Expression
}

func (segment Skip) New(config map[string]string) segments.Segment {
	var err error
	var skip uint = 1
	// Skip parameter is parsed but not validated, as that can only be done
	// in the context of a pipeline, i.e. when using Rewire.
	if parsedSkip, err := strconv.ParseUint(config["skip"], 10, 32); err == nil {
		skip = uint(parsedSkip)
	} else {
		if config["skip"] != "" {
			log.Println("[error] Skip: Could not parse 'skip' parameter, using default 1.")
		} else {
			log.Println("[info] Skip: 'skip' set to default 1.")
		}
	}

	invert, err := strconv.ParseBool(config["invert"])
	if err != nil {
		log.Println("[info] Skip: 'invert' set to default 'false'.")
	}

	newSegment := &Skip{
		Skip:      skip,
		Invert:    invert,
		Condition: config["condition"],
	}

	newSegment.expression, err = parser.Parse(config["condition"])
	if err != nil {
		log.Printf("[error] Skip: Syntax error in condition expression: %v", err)
		return nil
	}
	return newSegment
}

func (segment *Skip) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()

	filter := &visitors.Filter{}
	for msg := range segment.In {
		if match, err := filter.CheckFlow(segment.expression, msg); match != segment.Invert || err != nil {
			if err != nil {
				log.Printf("[error] FlowFilter: Semantic error in filter expression: %v", err)
				return // TODO: this will block the pipeline... find a way to tear down nicely
			}
			segment.AltOut <- msg
		} else {
			segment.Out <- msg
		}
	}
}

// Override default Rewire method for this segment.
func (segment *Skip) Rewire(chans []chan *flow.FlowMessage, in uint, out uint) {
	segment.In = chans[in]
	segment.Out = chans[out]
	if out+segment.Skip > uint(len(chans))-1 {
		log.Printf("[error] Skip: Configuration error, skipping the next %d segments is beyond the pipelines end, defaulting to 0.", segment.Skip)
		segment.Skip = 0
	}
	segment.AltOut = chans[out+segment.Skip]
}

func init() {
	segment := &Skip{}
	segments.RegisterSegment("skip", segment)
}
