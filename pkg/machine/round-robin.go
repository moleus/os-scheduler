package machine

import "fmt"

type RoundRobin struct {
	quantum int
}

func NewRoundRobin(quantum int) *RoundRobin {
	return &RoundRobin{quantum: quantum}
}

func (r *RoundRobin) ChooseToEvict(procs []*Process) []*Process {
	procsToEvict := make([]*Process, 0)
	for _, p := range procs {
		if p.runningTime > r.quantum {
			panic(fmt.Sprintf("Process %d has passed time %d, but quantum is %d", p.id, p.runningTime, r.quantum))
		}
		if p.runningTime == r.quantum || p.IsTaskCompleted() {
			procsToEvict = append(procsToEvict, p)
		}
	}
	return procsToEvict
}
