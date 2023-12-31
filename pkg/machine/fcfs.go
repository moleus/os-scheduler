package machine

type SelectionFIFO struct{}

// NonPreemptive - scheduler manages specific resource
type NonPreemptive struct{}

func NewNonPreemptive() Evictor {
	return &NonPreemptive{}
}

func (NonPreemptive) ChooseToEvict(procs []*Process) []*Process {
	procsToEvict := make([]*Process, 0)
	for _, p := range procs {
		if p.IsTaskCompleted() {
			procsToEvict = append(procsToEvict, p)
		}
	}
	return procsToEvict
}
