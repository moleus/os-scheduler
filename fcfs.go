package main

import (
	"fmt"
)

type DecisionMode int
const (
  NonPreemptive DecisionMode = iota
  Preemptive
)

type Scheduler interface {
  BeforeTick()
  AfterTick()
  Assign(p *Process)
  PushToQueue(p *Process)
}

type SelectionFunction interface {
  Select(queue *ProcQueue) (*Process, error)
}

type SelectionFIFO struct { }

func (s *SelectionFIFO) Select(queue *ProcQueue) (*Process, error) {
  return queue.Pop()
}

// FCFS - sheduler manages specific resource
type FCFS struct {
  name string
  resource Resourcer
  queue *ProcQueue
  decisionMode DecisionMode
  selectionFunc SelectionFunction
  clock GlobalTimer
}

func NewFCFS(name string, r Resourcer, clock GlobalTimer) *FCFS {
  return &FCFS{name: name, resource: r, queue: NewProcQueue(name, clock), decisionMode: NonPreemptive, selectionFunc: &SelectionFIFO{}, clock: clock}
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

func (f *FCFS) PushToQueue(p *Process) {
  f.queue.Push(p)
}
