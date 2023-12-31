package machine

import (
	"errors"
	"fmt"

	logger "github.com/Moleus/os-solver/pkg/logging"
)

type QueueElement struct {
	process   *Process
	enterTime int
}

type ProcQueue struct {
	name     string
	elements []QueueElement
	clock    logger.GlobalTimer
}

func NewProcQueue(name string, clock logger.GlobalTimer) *ProcQueue {
	return &ProcQueue{name, []QueueElement{}, clock}
}

func (pq *ProcQueue) Push(p *Process) {
	enterTime := pq.clock.GetCurrentTick()
	pq.elements = append(pq.elements, QueueElement{p, enterTime})
}

func (pq *ProcQueue) Pop() (*Process, error) {
	if len(pq.elements) == 0 {
		return &Process{}, errors.New("Queue is empty")
	}
	p := pq.elements[0]
	pq.elements = pq.elements[1:len(pq.elements)]
	return p.process, nil
}

func (pq *ProcQueue) Pick(process *Process) (*Process, error) {
	for i, p := range pq.elements {
		if p.process == process {
			pq.elements = append(pq.elements[:i], pq.elements[i+1:]...)
			return p.process, nil
		}
	}
	return &Process{}, errors.New(fmt.Sprintf("Process %d not found in queue %s", process.id, pq.name))
}

func (pq *ProcQueue) Len() int {
	return len(pq.elements)
}

func (pq *ProcQueue) GetQueueElements() []QueueElement {
	return pq.elements
}
