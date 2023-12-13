/*
FCFS scheduler for Single CPU
Each process has a sequence of CPU time and IO time switching
Each process has time it is added at start

We have N IO devices. Each have FCFS queue

Example input for proc1 and proc2 (CPU(x) means x time units of CPU time, IO(y) means y time units of IO time):
CPU(5) IO(2) CPU(1) IO(20) CPU(8)
CPU(4) IO(10) CPU(2)

Task:
1. measure time to complete all processes

- Процесс может в трех состояниях: CPU, IO, ready
- У нас 2 очереди: на CPU и на IO
- Мы не знаем, что будет дальше
- Процесс сам считает кол-во выполненных шагов
- Каждый квант времеи Планировщик смотрит только на наличие свободного места на CPU и IO и на очередь
*/
package main

import (
	"fmt"
)

type GlobalTimer interface {
  GetCurrentTick() int
}

type Machine struct {
  cpuScheduler Scheduler
  io1Scheduler Scheduler
  io2Scheduler Scheduler

  unscheduledProcs []Process
  clock *Clock
}

type Clock struct {
  currentTick int
}

func NewMachine(cpuScheduler Scheduler, io1Scheduler Scheduler, io2Scheduler Scheduler, clock *Clock) Machine {
  return Machine{cpuScheduler, io1Scheduler, io2Scheduler, []Process{}, clock}
}

func (c *Clock) GetCurrentTick() int {
  return c.currentTick
}

func (m *Machine) GetCurrentTick() int {
  return m.clock.currentTick
}

/*
0. Check completed and free + add them to waiting queue
1. Assign processes to CPU and IO
2. Increment counters
3. Set completed states
3. Debug output of current state
*/
// TODO: implmement completness checks and Queue management
// TODO: remove Tick, replace with IncrementCounters for all and UpdateState for running process
// TODO: add ready queue (cpu queue) and I/O queue (blocked state)
// TODO: add Preemt mechanism to stop process and move it to queue

func (m *Machine) allDone() bool {
  queuesAreEmpty := m.cpuScheduler.GetQueueLen() == 0 && m.io1Scheduler.GetQueueLen() == 0 && m.io2Scheduler.GetQueueLen() == 0
  unscheduledProcsIsEmpty := len(m.unscheduledProcs) == 0
  return queuesAreEmpty && unscheduledProcsIsEmpty
}

func (m *Machine) loop() {
  // infinite loop until all processes are done

  for {
    if m.allDone() {
      break
    }
    // If have waiting process in queue, assign it to CPU
    m.tick()
      // cpu, err := machine.GetFreeCpu()
      // if err != nil {
      //   fmt.Println(err)
      //   continue
      // }
      // // Assign process to CPU
      // machine.AssignToResource(&cpu, &cpuQ[0])
      // // Remove process from queue
      // cpuQ = cpuQ[1:]
    // }
  }
}

func (m *Machine) tick() {
  m.checkForNewProcs()

  m.cpuScheduler.CheckRunningProcs()
  m.io1Scheduler.CheckRunningProcs()
  m.io2Scheduler.CheckRunningProcs()

  m.handleAllEvictedProcs()

  m.cpuScheduler.ProcessQueue()
  m.io1Scheduler.ProcessQueue()
  m.io2Scheduler.ProcessQueue()

  m.clock.currentTick++
}

func (m *Machine) checkForNewProcs() {
  for i, p := range m.unscheduledProcs {
    if m.clock.currentTick < p.arrivalTime {
      // skip this proc. It's not time yet
      continue
    }
    fmt.Printf("Process %d arrived at %d\n", p.id, m.GetCurrentTick())
    m.cpuScheduler.PushToQueue(&p)
    // remove this proc from array
    m.unscheduledProcs = append(m.unscheduledProcs[:i], m.unscheduledProcs[i+1:]...)
  }
}

func (m *Machine) handleAllEvictedProcs() {
  ep := m.cpuScheduler.GetEvictedProcs()
  for _, p := range ep {
    m.handleEvictedProc(&p)
  }
}

func (m *Machine) handleEvictedProc(p *Process) {
  switch p.state {
  case TERMINATED:
    fmt.Printf("Process %d is done\n", p.id)
  case RUNNING | READY:
    // not finished or came from IO
    m.cpuScheduler.PushToQueue(p)
  case BLOCKED:
    m.pushToIO(p)
  case READS_IO:
    panic(fmt.Sprintf("Process %d evicted in READS_IO state but IO scheduler is nonpreemptive\n", p.id))
  }
}

func (m *Machine) pushToIO(p *Process) {
  switch p.NextTask().ResouceType {
  case IO1:
    m.io1Scheduler.PushToQueue(p)
  case IO2:
    m.io2Scheduler.PushToQueue(p)
  }
}

func (m *Machine) Run(processes []Process) {
  m.unscheduledProcs = processes

  m.loop()
}
