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
  resouceType ResourceType
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

func (p *Process) IsBlockedOrTerminated() bool {
  return p.state == BLOCKED || p.state == TERMINATED
}

func (p *Process) AssignToCpu() {
  p.state = RUNNING
  p.waitingTime = 0
  p.blockedTime = 0
}

func (p *Process) AssignToIo() {
  p.state = READS_IO
  p.waitingTime = 0
  p.blockedTime = 0
}

func (p *Process) Tick() {
  p.incrementCounters()
  p.updateState()
}

func (p *Process) incrementCounters() {
  switch p.state {
  case TERMINATED:
    fmt.Println("Process %d is already terminated", p.id)
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
    p.state = TERMINATED
    return
  }
  if p.CurTask().IsFinished() {
    p.completeTask()
  }
}

func (p *Process) completeTask() {
  p.currentTaskIndex++
  if p.currentTaskIndex > len(p.tasks) {
    p.state = TERMINATED
    return
  }
  switch p.CurTask().resouceType {
  case CPU:
    p.state = READY
  case IO:
    p.state = BLOCKED
  }
}

