package main

import (
	"fmt"

  "log/slog"
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
  ClearEvictedProcs()
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
    if p.IsTaskCompleted() {
      slog.Info(fmt.Sprintf("Evicting proc %d from resource %s\n", p.id, f.name))
      r.MustEvict(p)
      f.evictedProcs = append(f.evictedProcs, p)
    }
  }
}

func (f *FCFS) GetQueueLen() int {
  return len(f.queue.elements)
}

func (f *FCFS) assignFromQueue() {
  freeRes, err := f.resource.GetFree();
  if (err != nil) {
    slog.Debug(fmt.Sprintf("Resource %s is busy. Skipping scheduling\n", f.name))
    return
  }

  nextProc, err := f.selectionFunc.Select(f.queue)
  if (err != nil) {
    slog.Debug(fmt.Sprintf("No elements in queue %s\n", f.queue.name))
    return
  }

  err = freeRes.AssignToFree(nextProc)
  if (err != nil) {
    slog.Debug(fmt.Sprintf("Resource %s is busy. Skipping scheduling\n", f.name))
    return
  }
  slog.Info(fmt.Sprintf("Assigning process %d to resource %s\n", nextProc.id, f.name))
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

func (f *FCFS) ClearEvictedProcs() {
  f.evictedProcs = nil
}
