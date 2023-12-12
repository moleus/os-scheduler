package main

import (
	"fmt"
)

type DecisionMode int
const (
  NonPreemptive DecisionMode = iota
  Preemptive
)

type SelectionFunction interface {
  Select(queue ProcQueue) (*Process, error)
}

type SelectionFIFO struct { }

func (s *SelectionFIFO) Select(queue ProcQueue) (*Process, error) {
  return queue.Pop()
}

// FCFS - sheduler manages specific resource
type FCFS struct {
  r Resourcer
  queue ProcQueue
  decisionMode DecisionMode
  selectionFunc SelectionFunction
}

func NewFCFS(r Resourcer) *FCFS {
  return &FCFS{r: r}
}

func (f *FCFS) Assign(p *Process) {
  f.r.AssignToFree(p)
}

func (f *FCFS) Tick() {
  nextProc, err := f.selectionFunc.Select(f.queue)
  if (err != nil) {
    fmt.Println("No elements in queue %s", f.queue.name)
  }
  f.Assign(nextProc)

  // iterate over
}
