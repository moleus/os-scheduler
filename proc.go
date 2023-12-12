package main

import "fmt"

type ProcState int
const (
  READY ProcState = iota  // ready to run on CPU
  RUNNING  // runs on CPU
  BLOCKED  // waits or reads from IO
  TERMINATED  // completed
)

type Task struct {
  resouceType ResourceType
  passedTime int
  time int
}

type Process struct {
  id int
  arrivalTime int
  state ProcState

  currentTaskIndex int
  tasks []Task

  waitingTime int
  cpuTimePassed int
  cpuTimeTotal int

  ioTimePassed int
  ioTimeTotal int
}

func (p *Process) EstimatedTaskTime() int {
  return p.tasks[p.currentTaskIndex].time
}

func (t *Task) IsFinished() bool {
  return t.passedTime == t.time
}

func (p *Process) CurTask() *Task {
  return &p.tasks[p.currentTaskIndex]
}

type Tickable interface {
  Tick()
}

func (p *Process) IncrementCounters() {
  switch p.state {
  case DONE:
    fmt.Println("Process %d is already done", p.id)
  case READY:
    p.waitingTime++
  case RUNNING:
    p.CurTask().passedTime++
    if p.CurTask().resouceType == CPU {
      p.cpuTimePassed++
    } else {
      p.ioTimePassed++
    }
  }

  if p.CurTask().passedTime > p.CurTask().time {
    panic(fmt.Sprintf("Passed time is greater than total time for proc %d, Task %d", p.id, p.currentTaskIndex))
  }
}

func (p *Process) UpdateState() {
  if p.currentTaskIndex == len(p.tasks) {
    p.state = DONE
    return
  }
  if p.CurTask().IsFinished() {
    p.state = READY
  } else {
    p.state = RUNNING
  }
}
