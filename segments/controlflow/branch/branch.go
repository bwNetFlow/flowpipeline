package branch

import (
	"log"
	"sync"

	"github.com/bwNetFlow/flowpipeline/segments"
	flow "github.com/bwNetFlow/protobuf/go"
)

// This mirrors the proper implementation in the pipeline package. This
// duplication is to avoid the import cycle.
type Pipeline interface {
	Start()
	Close()
	GetInput() chan *flow.FlowMessage
	GetOutput() <-chan *flow.FlowMessage
	GetDrop() <-chan *flow.FlowMessage
	IsEmpty() bool
}

type Branch struct {
	segments.BaseSegment
	condition   Pipeline
	then_branch Pipeline
	else_branch Pipeline
}

func (segment Branch) New(config map[string]string) segments.Segment {
	return &Branch{}
}

func (segment *Branch) ImportBranches(condition interface{}, then_branch interface{}, else_branch interface{}) {
	segment.condition = condition.(Pipeline)
	segment.then_branch = then_branch.(Pipeline)
	segment.else_branch = else_branch.(Pipeline)
}

func (segment *Branch) Run(wg *sync.WaitGroup) {
	defer func() {
		if !segment.condition.IsEmpty() {
			segment.condition.Close()
		}
		if !segment.then_branch.IsEmpty() {
			segment.then_branch.Close()
		}
		if !segment.else_branch.IsEmpty() {
			segment.else_branch.Close()
		}
		close(segment.Out)
		wg.Done()
	}()

	// This large branching structures covers any case of defined and
	// undefined branches and or conditions and links the appropriate
	// channels accordingly while providing user facing log messages on
	// what it is doing.
	// The alternativ is to replace empty configurations of any conditional
	// or branch with a hardcoded `pass` segment, which would allow us to
	// do only the very first for loop in this structure and be done with
	// it.
	if !segment.condition.IsEmpty() {
		go segment.condition.Start()
		log.Println("[info] Branch: Started subpipeline for `if` segments.")
		if !segment.then_branch.IsEmpty() && !segment.else_branch.IsEmpty() {
			go segment.then_branch.Start()
			log.Println("[info] Branch: Started subpipeline for `then` segments.")
			go segment.else_branch.Start()
			log.Println("[info] Branch: Started subpipeline for `else` segments.")
			for pre := range segment.In {
				segment.condition.GetInput() <- pre
				select {
				case msg := <-segment.condition.GetOutput():
					segment.then_branch.GetInput() <- msg
					segment.Out <- <-segment.then_branch.GetOutput()
				case msg := <-segment.condition.GetDrop():
					segment.else_branch.GetInput() <- msg
					segment.Out <- <-segment.else_branch.GetOutput()
				}
			}
		} else if !segment.then_branch.IsEmpty() {
			log.Println("[info] Branch: Empty else branch, fast-forwarding drops in conditional to the next segment.")
			go segment.then_branch.Start()
			log.Println("[info] Branch: Started subpipeline for `then` segments.")
			for pre := range segment.In {
				segment.condition.GetInput() <- pre
				select {
				case msg := <-segment.condition.GetOutput():
					segment.then_branch.GetInput() <- msg
					segment.Out <- <-segment.then_branch.GetOutput()
				case msg := <-segment.condition.GetDrop():
					segment.Out <- msg
				}
			}
		} else if !segment.else_branch.IsEmpty() {
			log.Println("[info] Branch: Empty then branch, fast-forwarding matches in conditional to the next segment.")
			go segment.else_branch.Start()
			log.Println("[info] Branch: Started subpipeline for `else` segments.")
			for pre := range segment.In {
				segment.condition.GetInput() <- pre
				select {
				case msg := <-segment.condition.GetOutput():
					segment.Out <- msg
				case msg := <-segment.condition.GetDrop():
					segment.else_branch.GetInput() <- msg
					segment.Out <- <-segment.else_branch.GetOutput()
				}
			}
		}
	} else {
		if !segment.else_branch.IsEmpty() {
			log.Println("[warning] Branch: Empty conditional, else branch is unreachable.")
		}
		if !segment.then_branch.IsEmpty() {
			log.Println("[info] Branch: Empty conditional, acting as an unconditional branch.")
			go segment.then_branch.Start()
			log.Println("[info] Branch: Started subpipeline for `then` segments.")
			for pre := range segment.In {
				segment.then_branch.GetInput() <- pre
				segment.Out <- <-segment.then_branch.GetOutput()
			}
		} else {
			log.Println("[warning] Branch: Empty conditional and empty `then` branch, acting as `pass` segment.")
			for pre := range segment.In {
				segment.Out <- pre
			}
		}
	}
}

// Override default Rewire method for this segment.
func (segment *Branch) Rewire(chans []chan *flow.FlowMessage, in uint, out uint) {
	segment.In = chans[in]
	segment.Out = chans[out]
}

func init() {
	segment := &Branch{}
	segments.RegisterSegment("branch", segment)
}
