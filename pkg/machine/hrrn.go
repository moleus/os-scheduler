package machine

import "errors"

type SelectionHRRN struct{}

func NewSelectionHRRN() SelectionFunction {
	return &SelectionHRRN{}
}

// Select - picks process with the shortest remaining time
func (s SelectionHRRN) Select(queue *ProcQueue) (*Process, error) {
	elements := queue.GetQueueElements()
	if len(elements) == 0 {
		return &Process{}, errors.New("queue is empty")
	}
	proc := getMinByHighestResponseRatio(elements)
	return queue.Pick(proc)
}

func getMinByHighestResponseRatio(elements []QueueElement) *Process {
	// response ratio = (wait time + service time) / service time

	minProc := elements[0].process
	minResponseRatio := elements[0].process.waitingTime / elements[0].process.CurTask().TotalTime

	for _, qe := range elements {
		p := qe.process
		responseRatio := p.waitingTime / p.CurTask().TotalTime
		if responseRatio < minResponseRatio {
			minProc = p
			minResponseRatio = responseRatio
		}
	}

	return minProc
}
