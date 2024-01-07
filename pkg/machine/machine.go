/*
NonPreemptive scheduler for Single CPU
Each process has a sequence of CPU time and IO time switching
Each process has time it is added at start

We have N IO devices. Each have NonPreemptive queue

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
package machine

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"
)

type SnapshotStateFunc func(row string)

type Machine struct {
	cpuScheduler Scheduler
	io1Scheduler Scheduler
	io2Scheduler Scheduler

	unscheduledProcs  []*Process
	runningProcs      []*Process
	clock             *Clock
	logger            *slog.Logger
	snapshotStateFunc SnapshotStateFunc
	cpuCount          int
}

type Clock struct {
	CurrentTick int
}

func NewMachine(cpuScheduler Scheduler, io1Scheduler Scheduler, io2Scheduler Scheduler, clock *Clock, logger *slog.Logger, snapshotStateFunc SnapshotStateFunc, cpuCount int) Machine {
	return Machine{cpuScheduler, io1Scheduler, io2Scheduler, []*Process{}, []*Process{}, clock, logger, snapshotStateFunc, cpuCount}
}

func (c *Clock) GetCurrentTick() int {
	return c.CurrentTick
}

func (m *Machine) GetCurrentTick() int {
	return m.clock.CurrentTick
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
	return len(m.runningProcs) == 0 && len(m.unscheduledProcs) == 0
}

func (m *Machine) loop() {
	for {
		if m.allDone() {
			break
		}
		m.tick()
	}
}

func (m *Machine) tick() {
	// problem: evicted process comes before new process?
	m.checkForNewProcs()

	m.cpuScheduler.CheckRunningProcs()
	m.io1Scheduler.CheckRunningProcs()
	m.io2Scheduler.CheckRunningProcs()

	m.handleAllEvictedProcs()

	m.cpuScheduler.ProcessQueue()
	m.io1Scheduler.ProcessQueue()
	m.io2Scheduler.ProcessQueue()

	m.snapshotStateFunc(m.dumpState())

	m.clock.CurrentTick++

	for _, p := range m.runningProcs {
		p.Tick()
	}
}

func (m *Machine) checkForNewProcs() {
	unscheduleCandidates := make([]*Process, 0)
	for _, p := range m.unscheduledProcs {
		if m.clock.CurrentTick < p.arrivalTime {
			// skip this proc. It's not time yet
			continue
		}
		m.logger.Info(fmt.Sprintf("Process %d arrived at tick %d", p.id, m.GetCurrentTick()))
		m.cpuScheduler.PushToQueue(p)
		m.runningProcs = append(m.runningProcs, p)
		unscheduleCandidates = append(unscheduleCandidates, p)
	}
	m.unscheduledProcs = slices.DeleteFunc(m.unscheduledProcs, func(p *Process) bool {
		return slices.Contains(unscheduleCandidates, p)
	})
	m.logger.Debug(fmt.Sprintf("Unscheduled procs: %d", len(m.unscheduledProcs)))
}

func (m *Machine) handleAllEvictedProcs() {
	ep := m.cpuScheduler.GetEvictedProcs()
	ep = append(ep, m.io1Scheduler.GetEvictedProcs()...)
	ep = append(ep, m.io2Scheduler.GetEvictedProcs()...)
	for _, p := range ep {
		m.handleEvictedProc(p)
	}
	m.cpuScheduler.ClearEvictedProcs()
	m.io1Scheduler.ClearEvictedProcs()
	m.io2Scheduler.ClearEvictedProcs()
}

func (m *Machine) handleEvictedProc(p *Process) {
	switch p.state {
	case TERMINATED:
		m.logger.Info(fmt.Sprintf("Process %d is done at tick %d", p.id, m.GetCurrentTick()))
		// remove from running procs
		for i, rp := range m.runningProcs {
			if rp.id == p.id {
				m.runningProcs = append(m.runningProcs[:i], m.runningProcs[i+1:]...)
				break
			}
		}
	case RUNNING, READY:
		// not finished or came from IO
		m.cpuScheduler.PushToQueue(p)
	case BLOCKED:
		m.pushToIO(p)
	case READS_IO:
		panic(fmt.Sprintf("Process %d evicted in READS_IO state but IO scheduler is nonpreemptive", p.id))
	}
}

func (m *Machine) pushToIO(p *Process) {
	switch p.CurTask().ResouceType {
	case IO1:
		m.logger.Debug(fmt.Sprintf("Process %d is blocked on IO1", p.id))
		m.io1Scheduler.PushToQueue(p)
	case IO2:
		m.logger.Debug(fmt.Sprintf("Process %d is blocked on IO2", p.id))
		m.io2Scheduler.PushToQueue(p)
	case CPU:
		panic(fmt.Sprintf("Proc %d is blocked by current task is cpu", p.id))
	}
}

// DumpState - prints running processes on each cpu and io in one line
// output format:
// {tick} {procid on first cpu} {procid on second cpu} ... {procid on last cpu} {procid on io1} {procid on io2}
// if no proc on cpu or io, output - instead of id
func (m *Machine) dumpState() string {
	cpusStateString := make([]string, m.cpuCount)

	cpus := m.cpuScheduler.GetResource().(*CpuPool).cpus
	for i, cpu := range cpus {
		cpusStateString[i] = resourceStateToString(cpu)
	}
	cpusString := strings.Join(cpusStateString, " ")

	io1 := m.io1Scheduler.GetResource().(*Resource)
	io1State := resourceStateToString(io1)

	io2 := m.io2Scheduler.GetResource().(*Resource)
	io2State := resourceStateToString(io2)

	return fmt.Sprintf("%3d | %s | %s %s", m.GetCurrentTick(), cpusString, io1State, io2State)
}

func resourceStateToString(r *Resource) string {
	if r.state == BUSY {
		return fmt.Sprintf("%d", r.currentProc.id+1)
	}
	return "-"
}

func (m *Machine) Run(processes []*Process) {
	m.unscheduledProcs = make([]*Process, len(processes))
	copy(m.unscheduledProcs, processes)

	m.loop()
}
