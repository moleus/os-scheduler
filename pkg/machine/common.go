package machine

import (
	"fmt"
	log "github.com/Moleus/os-solver/pkg/logging"
	"log/slog"
)

type Scheduler interface {
	CheckRunningProcs()
	ProcessQueue()
	Assign(p *Process)
	PushToQueue(p *Process)
	GetQueueLen() int
	GetEvictedProcs() []*Process
	ClearEvictedProcs()
	GetResource() Resourcer
}

type SelectionFunction interface {
	Select(queue *ProcQueue) (*Process, error)
}

func NewSelectionFIFO() *SelectionFIFO {
	return &SelectionFIFO{}
}

func (s *SelectionFIFO) Select(queue *ProcQueue) (*Process, error) {
	return queue.Pop()
}

type Evictor interface {
	ChooseToEvict(procs []*Process) []*Process
}

type SchedulerWrapper struct {
	name string

	resource Resourcer
	queue    *ProcQueue
	clock    log.GlobalTimer

	evictedProcs []*Process
	logger       *slog.Logger

	selectionFunc SelectionFunction
	evictor       Evictor
}

func NewSchedulerWrapper(name string, selection SelectionFunction, evictor Evictor, r Resourcer, clock log.GlobalTimer, logger *slog.Logger) *SchedulerWrapper {
	queue := NewProcQueue(name, clock)
	evictedProcs := make([]*Process, 0)
	return &SchedulerWrapper{name: name, resource: r, queue: queue, clock: clock, evictedProcs: evictedProcs, logger: logger, selectionFunc: selection, evictor: evictor}
}

func (b *SchedulerWrapper) CheckRunningProcs() {
	procsToEvict := b.evictor.ChooseToEvict(b.resource.GetProcs())
	for _, p := range procsToEvict {
		b.logger.Info(fmt.Sprintf("Evicting process %d from resource %s", p.id, b.name))
		b.resource.MustEvict(p)
		b.evictedProcs = append(b.evictedProcs, p)
	}
}

func (b *SchedulerWrapper) ProcessQueue() {
	b.assignFromQueue()
}

func (b *SchedulerWrapper) Assign(p *Process) {
	b.resource.AssignToFree(p)
}

func (b *SchedulerWrapper) PushToQueue(p *Process) {
	b.queue.Push(p)
}

func (b *SchedulerWrapper) GetQueueLen() int {
	return b.queue.Len()
}

func (b *SchedulerWrapper) GetEvictedProcs() []*Process {
	return b.evictedProcs
}

func (b *SchedulerWrapper) ClearEvictedProcs() {
	b.evictedProcs = []*Process{}
}

func (b *SchedulerWrapper) GetResource() Resourcer {
	return b.resource
}

func (b *SchedulerWrapper) assignFromQueue() {
	freeRes, err := b.resource.GetFree()
	if err != nil {
		b.logger.Debug(fmt.Sprintf("Resource %s is busy. Skipping scheduling", b.name))
		return
	}

	nextProc, err := b.selectionFunc.Select(b.queue)
	if err != nil {
		b.logger.Debug(fmt.Sprintf("No elements in queue %s", b.queue.name))
		return
	}

	err = freeRes.AssignToFree(nextProc)
	if err != nil {
		b.logger.Debug(fmt.Sprintf("Resource %s is busy. Skipping scheduling", b.name))
		return
	}
	b.logger.Info(fmt.Sprintf("Assigning process %d to resource %s", nextProc.id, b.name))
}
