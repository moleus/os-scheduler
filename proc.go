package main

import "fmt"

type ProcState int
const (
  READY ProcState = iota  // ready to run on CPU
  RUNNING  // runs on CPU
  BLOCKED  // waits for IO
  READS_IO // reads from IO
  TERMINATED  // completed
)

type Task struct {
  ResouceType ResourceType
  passedTime int
  totalTime int
}

type Process struct {
  id int
  arrivalTime int
  state ProcState

  currentTaskIndex int
  tasks []Task

  waitingTime int
  blockedTime int
}

func (p *Process) EstimatedTaskTime() int {
  return p.tasks[p.currentTaskIndex].totalTime
}

func (t *Task) IsFinished() bool {
  return t.passedTime == t.totalTime
}

func (p *Process) CurTask() *Task {
  return &p.tasks[p.currentTaskIndex]
}

func (p *Process) NextTask() *Task {
  return &p.tasks[p.currentTaskIndex+1]
}

func (p *Process) IsBlockedOrTerminated() bool {
  return p.state == BLOCKED || p.state == TERMINATED
}

func (p *Process) AssignToCpu() {
  p.state = RUNNING
  p.waitingTime = 0
  p.blockedTime = 0
}

func (p *Process) AssignToIo() {
  fmt.Printf("Process %d assigned to IO\n", p.id)
  p.state = READS_IO
  p.waitingTime = 0
  p.blockedTime = 0
}

func (p *Process) Tick() {
  p.incrementCounters()
  p.updateState()
}

func (p *Process) incrementCounters() {
  fmt.Printf("Process %d ticked. State: %v\n", p.id, p.state)
  switch p.state {
  case TERMINATED:
    fmt.Printf("Process %d is already terminated\n", p.id)
  case READY:
    p.waitingTime++
  case BLOCKED:
    p.blockedTime++
  case RUNNING:
    p.CurTask().passedTime++
  }

  if p.CurTask().passedTime > p.CurTask().totalTime {
    panic(fmt.Sprintf("Passed time is greater than total time for proc %d, Task %d", p.id, p.currentTaskIndex))
  }
}

func (p *Process) updateState() {
  if p.currentTaskIndex == len(p.tasks) {
    fmt.Printf("Process %d finished all tasks\n", p.id)
    p.state = TERMINATED
    return
  }
  if p.CurTask().IsFinished() {
    fmt.Printf("Process %d finished task %d\n", p.id, p.currentTaskIndex)
    p.completeTask()
  }
}

func (p *Process) completeTask() {
  p.currentTaskIndex++
  if p.currentTaskIndex > len(p.tasks) {
    p.state = TERMINATED
    return
  }
  switch p.CurTask().ResouceType {
  case CPU:
    p.state = READY
  case IO1 | IO2:
    p.state = BLOCKED
  }
}

