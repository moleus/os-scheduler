package machine

import (
	"errors"
	"slices"
	"sort"
)

// TODO: add list of evicted processes to distinguish between evicted and new comming processes
type SchedulerSRT struct {
	oldProcs []*Process
	procQueue    *ProcQueue
	cpuCount     int
}

func NewSchedulerSRT(procQueue *ProcQueue, cpuCount int) *SchedulerSRT {
	evictedProcs := make([]*Process, 0)
	return &SchedulerSRT{evictedProcs, procQueue, cpuCount}
}

// Select - picks process with the shortest remaining time
func (s *SchedulerSRT) Select(queue *ProcQueue) (*Process, error) {
	elements := queue.GetQueueElements()
	if len(elements) == 0 {
		return &Process{}, errors.New("queue is empty")
	}

	// if doesn't contain new procs, then return
	hasNewProcs := hasNewProcs(queue, s.oldProcs)

	if !hasNewProcs {
		return &Process{}, errors.New("no new procs")
	}

	proc := getMinByRemainingTaskTime(elements)

	return queue.Pick(proc)
}

func hasNewProcs(queue *ProcQueue, oldProcs []*Process) bool {
	elements := queue.GetQueueElements()
	if len(elements) == 0 {
		return false
	}

	for _, e := range elements {
		if !slices.Contains(oldProcs, e.process) {
			return true
		}
	}
	return false
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

func (s *SchedulerSRT) ChooseToEvict(procs []*Process) []*Process {
	procsToEvict := make([]*Process, 0)
	freeCpus := s.cpuCount - len(procs)

	// first evict completed procs
	for _, p := range procs {
		if p.IsTaskCompleted() {
			s.oldProcs = slices.DeleteFunc(s.oldProcs, func(evicted *Process) bool {
				return p == evicted
			})
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

	// мы убираем из очереди только когда в очереди есть новый процесс. Если в очереди нет новых процессов, то никого не убираем
	if !hasNewProcs(s.procQueue, s.oldProcs) {
		return procsToEvict
	}

	queueElements = slices.DeleteFunc(queueElements, func(qe QueueElement) bool {
		return slices.Contains(s.oldProcs, qe.process)
	})

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
			procsToEvict = append(procsToEvict, procsCopy[c])
			c++
			q++
		} else {
			c++
		}
	}

	return procsToEvict
}
