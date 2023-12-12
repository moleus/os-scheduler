package main

type DecisionMode int
const (
  NonPreemptive DecisionMode = iota
  Preemptive
)

// FCFS - sheduler manages specific resource
type FCFS struct {
  r Resourcer
  decisionMode DecisionMode
}

type SelectionFunction interface {
  Select(queue []Process) *Process
}

func NewFCFS(r Resourcer) *FCFS {
  return &FCFS{r: r}
}

func (f *FCFS) Assign(p *Process) {
  f.r.AssignToFree(p)
}

func (fcfs *FCFS) Tick() {
  // iterate over
}
