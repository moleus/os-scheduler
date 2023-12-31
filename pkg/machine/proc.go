package machine

import (
	"fmt"
	"log/slog"
)

type ProcState int

const (
	READY      ProcState = iota // ready to run on CPU
	RUNNING                     // runs on CPU
	BLOCKED                     // waits for IO
	READS_IO                    // reads from IO
	TERMINATED                  // completed
)

type Task struct {
	ResouceType ResourceType
	passedTime  int
	TotalTime   int
}

type ProcStatistics interface {
	GetStats() ProcStats
}

type ProcStats struct {
	ProcId             int
	EntranceTime       int
	ServiceTime        int
	ExitTime           int
	StartTime          int
	ReadyOrBlockedTime int
	TurnaroundTime     int
}

type Process struct {
	id          int
	arrivalTime int
	state       ProcState

	currentTaskIndex int
	tasks            []Task

	waitingTime int
	blockedTime int
	runningTime int

	logger *slog.Logger

	procStats *ProcStats
}

func NewProcess(id int, arrivalTime int, tasks []Task, logger *slog.Logger) *Process {
	procStats := &ProcStats{ProcId: id, EntranceTime: arrivalTime, StartTime: -1}
	return &Process{id, arrivalTime, READY, 0, tasks, 0, 0, 0, logger, procStats}
}

func (p *Process) GetStats() ProcStats {
	return *p.procStats
}

func (p *Process) EstimatedTaskTime() int {
	return p.tasks[p.currentTaskIndex].TotalTime
}

func (t *Task) IsFinished() bool {
	return t.passedTime == t.TotalTime
}

func (p *Process) CurTask() *Task {
	return &p.tasks[p.currentTaskIndex]
}

func (p *Process) TaskRemainingTime() int {
	return p.CurTask().TotalTime - p.CurTask().passedTime
}

func (p *Process) NextTask() *Task {
	return &p.tasks[p.currentTaskIndex+1]
}

func (p *Process) IsTaskCompleted() bool {
	return p.state == BLOCKED || p.state == TERMINATED || p.state == READY
}

func (p *Process) AssignToCpu() {
	p.state = RUNNING
	p.waitingTime = 0
	p.blockedTime = 0
}

func (p *Process) AssignToIo() {
	p.logger.Info(fmt.Sprintf("Process %d assigned to IO", p.id))
	p.state = READS_IO
	p.waitingTime = 0
	p.blockedTime = 0
}

func (p *Process) Tick() {
	p.incrementCounters()
	p.updateState()
	p.updateGlobalProcStatsOnTick()
}

func (p *Process) updateGlobalProcStatsOnTick() {
	if p.state == RUNNING || p.state == READS_IO {
		if p.procStats.StartTime == -1 {
			p.procStats.StartTime = p.arrivalTime + p.waitingTime
		}
		p.procStats.ServiceTime++
	} else if p.state == READY || p.state == BLOCKED {
		p.procStats.ReadyOrBlockedTime++
	} else if p.state == TERMINATED {
		p.procStats.TurnaroundTime = p.procStats.ServiceTime + p.procStats.ReadyOrBlockedTime
		p.procStats.ExitTime = p.procStats.TurnaroundTime + p.procStats.EntranceTime
	}
}

func (p *Process) incrementCounters() {
	p.logger.Debug(fmt.Sprintf("Process %d ticked. State: %v", p.id, p.state))
	switch p.state {
	case TERMINATED:
		p.logger.Warn(fmt.Sprintf("Process %d is already terminated", p.id))
	case READY:
		p.waitingTime++
	case BLOCKED:
		p.blockedTime++
	case RUNNING, READS_IO:
		p.runningTime++
		p.CurTask().passedTime++
	}

	if p.CurTask().passedTime > p.CurTask().TotalTime {
		panic(fmt.Sprintf("Passed time is greater than total time for proc %d, Task %d", p.id, p.currentTaskIndex))
	}
}

func (p *Process) updateState() {
	if p.currentTaskIndex == len(p.tasks) {
		p.logger.Info(fmt.Sprintf("Process %d finished all tasks", p.id))
		p.state = TERMINATED
		return
	}
	if p.CurTask().IsFinished() {
		p.logger.Debug(fmt.Sprintf("Process %d finished task %d", p.id, p.currentTaskIndex))
		p.completeTask()
	}
}

func (p *Process) completeTask() {
	p.runningTime = 0
	p.currentTaskIndex++
	if p.currentTaskIndex >= len(p.tasks) {
		p.state = TERMINATED
		return
	}
	switch p.CurTask().ResouceType {
	case CPU:
		p.state = READY
		p.logger.Debug(fmt.Sprintf("Process %d ready", p.id))
	case IO1, IO2:
		p.state = BLOCKED
		p.logger.Debug(fmt.Sprintf("Process %d blocked on IO%d", p.id, p.CurTask().ResouceType-IO1+1))
	}
}

func (p *Process) onEvict() {
	p.logger.Debug(fmt.Sprintf("Process %d evicted", p.id))

	p.runningTime = 0
	switch p.state {
	case RUNNING:
		p.state = READY
	case READS_IO:
		p.state = BLOCKED
		panic(fmt.Sprintf("Process %d evicted in READS_IO state but IO scheduler is nonpreemptive", p.id))
	}
}
