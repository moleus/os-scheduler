package machine

type SelectionFIFO struct{}

// FCFS - scheduler manages specific resource
type FCFS struct{}

func NewFCFS() *FCFS {
	return &FCFS{}
}

func (f *FCFS) ChooseToEvict(procs []*Process) []*Process {
	procsToEvict := make([]*Process, 0)
	for _, p := range procs {
		if p.IsTaskCompleted() {
			procsToEvict = append(procsToEvict, p)
		}
	}
	return procsToEvict
}
