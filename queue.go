package main

import (
	"errors"
)

type QueueElement struct {
  process *Process
  enterTime int
}

type ProcQueue struct {
  name string
  elements []QueueElement
  clock GlobalTimer
}

func NewProcQueue(name string, clock GlobalTimer) *ProcQueue {
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
