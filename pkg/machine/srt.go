package machine

import (
	"errors"
	"slices"
	"sort"
)

type SelectionSRT struct{}

func NewSelectionSRT() SelectionFunction {
	return &SelectionSRT{}
}

// Select - picks process with the shortest remaining time
func (s SelectionSRT) Select(queue *ProcQueue) (*Process, error) {
	elements := queue.GetQueueElements()
	if len(elements) == 0 {
		return &Process{}, errors.New("queue is empty")
	}
	proc := getMinByRemainingTaskTime(elements)
	return queue.Pick(proc)
}

func getMinByRemainingTaskTime(elements []QueueElement) *Process {
	minProc := elements[0].process
	minTimeLeft := minProc.CurTask().TotalTime - minProc.CurTask().passedTime
	for _, qe := range elements {
		p := qe.process
		timeLeft := p.CurTask().TotalTime - p.CurTask().passedTime
		if timeLeft < minTimeLeft {
			minProc = p
			minTimeLeft = timeLeft
		}
	}
	return minProc
}

// SrtEvictor - Shortest Remaining Time (SRT) scheduler
// First, we evict process from CPU if we have process in queue shorter than current
// I think, we should evict as many processes as we can, so for each proc in Queue we check if it's shorter than running processes
// Algorithm:
// sort processes in queue by remaining time (increase)
// sort processes in CPU by remaining time (increase)
// if queue[q] < cpu[c] then evict cpu[c] and q++, c++
// if queue[q] > cpu[c] then return (we can't evict anything)
// if queue[q] == cpu[c] then evict cpu[0] and q++, c++
type SrtEvictor struct {
	procQueue *ProcQueue
	cpuCount  int
}

func NewSRTEvictor(procQueue *ProcQueue, cpuCount int) Evictor {
	return &SrtEvictor{procQueue: procQueue, cpuCount: cpuCount}
}

func (s *SrtEvictor) ChooseToEvict(procs []*Process) []*Process {
	procsToEvict := make([]*Process, 0)
	freeCpus := s.cpuCount - len(procs)

	// first evict completed procs
	for _, p := range procs {
		if p.IsTaskCompleted() {
			procsToEvict = append(procsToEvict, p)
			freeCpus++
		}
	}

	srcQueueElements := s.procQueue.GetQueueElements()
	if len(srcQueueElements) == 0 {
		return procsToEvict
	}

	queueElements := make([]QueueElement, len(srcQueueElements))
	copy(queueElements, srcQueueElements)

	sort.Slice(queueElements, func(i, j int) bool {
		return queueElements[i].process.TaskRemainingTime() < queueElements[j].process.TaskRemainingTime()
	})

	// TODO: check that we copy pointers not values
	procsCopy := make([]*Process, len(procs))
	copy(procsCopy, procs)

	sort.Slice(procsCopy, func(i, j int) bool {
		return procsCopy[i].TaskRemainingTime() < procsCopy[j].TaskRemainingTime()
	})

	for q, c := 0, 0; q < len(queueElements) && c < len(procsCopy); {
		if freeCpus > 0 {
			// don't evict shortest proc because we have a free cpu for it
			freeCpus--
			c++
			continue
		}
		if slices.Contains(procsToEvict, procsCopy[c]) {
			c++
			continue
		}
		if queueElements[q].process.TaskRemainingTime() < procsCopy[c].TaskRemainingTime() {
			c++
			q++
		} else {
			q++
		}
	}

	return procsToEvict
}
