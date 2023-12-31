package machine

import (
	"errors"

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

func (pq *ProcQueue) Len() int {
	return len(pq.elements)
}
