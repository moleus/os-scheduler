package machine

import "errors"

type SelectionSPN struct{}

func NewSelectionSPN() SelectionFunction {
	return &SelectionSPN{}
}

func (s *SelectionSPN) Select(queue *ProcQueue) (*Process, error) {
	elements := queue.GetQueueElements()
	if len(elements) == 0 {
		return &Process{}, errors.New("queue is empty")
	}
	proc := getMinByTaskTime(elements)
	return queue.Pick(proc)
}

func getMinByTaskTime(elements []QueueElement) *Process {
	minProc := elements[0].process
	for _, qe := range elements {
		p := qe.process
		// TODO: maybe we need to compare by time left?
		if p.CurTask().TotalTime < minProc.CurTask().TotalTime {
			minProc = p
		}
	}
	return minProc
}
