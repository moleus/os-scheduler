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
  resource Resourcer
  queue ProcQueue
  decisionMode DecisionMode
  selectionFunc SelectionFunction
}

func NewFCFS(r Resourcer) *FCFS {
  return &FCFS{resource: r}
}

func (f *FCFS) Assign(p *Process) {
  f.resource.AssignToFree(p)
}

func (f *FCFS) evictTerminatedProcs() {
  r := f.resource
  procs := r.GetProcs()
  for _, p := range procs {
    if p.IsBlockedOrTerminated() {
      r.MustEvict(&p)
    }
  }
}

func (f *FCFS) Tick() {
  f.evictTerminatedProcs()

  freeRes, err := f.resource.GetFree();
  if (err != nil) {
    fmt.Println("Resource is bussy. Skipping scheduling")
    return
  }

  nextProc, err := f.selectionFunc.Select(f.queue)
  if (err != nil) {
    fmt.Println("No elements in queue %s", f.queue.name)
    return
  }

  freeRes.AssignToFree(nextProc)

  // iterate over
}
