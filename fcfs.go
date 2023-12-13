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
  CheckRunningProcs()
  ProcessQueue()
  Assign(p *Process)
  PushToQueue(p *Process)
  GetQueueLen() int
  GetEvictedProcs() []*Process
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

  evictedProcs []*Process
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
      f.evictedProcs = append(f.evictedProcs, &p)
    }
  }
}

func (f *FCFS) GetQueueLen() int {
  return len(f.queue.elements)
}

func (f *FCFS) assignFromQueue() {
  f.evictTerminatedProcs()

  freeRes, err := f.resource.GetFree();
  if (err != nil) {
    fmt.Println("Resource is bussy. Skipping scheduling")
    return
  }

  nextProc, err := f.selectionFunc.Select(f.queue)
  if (err != nil) {
    fmt.Printf("No elements in queue %s\n", f.queue.name)
    return
  }

  freeRes.AssignToFree(nextProc)
  // TODO: update proc state
}

func (f *FCFS) CheckRunningProcs() {
  f.evictTerminatedProcs()
}

func (f *FCFS) ProcessQueue() {
  f.assignFromQueue()
}

func (f *FCFS) PushToQueue(p *Process) {
  f.queue.Push(p)
}

func (f *FCFS) GetEvictedProcs() []*Process {
  return f.evictedProcs
}
