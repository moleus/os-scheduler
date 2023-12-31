package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"

	log "github.com/Moleus/os-solver/pkg/logging"
	m "github.com/Moleus/os-solver/pkg/machine"
)

var (
	cpuCount      = flag.Int("cpus", 4, "Number of CPUs")
	inputFile     = flag.String("input", "", "Input file")
	outputFile    = flag.String("output", "result.txt", "Output file")
	procStatsFile = flag.String("procStats", "procStats.txt", "Process stats file")
	schedAlgo     = flag.String("sched", "fcfs", "Scheduling algorithm (default: fcfs). Possible values: fcfs, rr1, rr2, spn, srt, hrrn")
)

func calcArrivalTime(procId int) int {
	return procId * 2
}

func ParseTask(task string) m.Task {
	task = strings.TrimSpace(task)
	var taskType m.ResourceType
	taskTypeStr := task[:3]
	switch taskTypeStr {
	case "IO1":
		taskType = m.IO1
	case "IO2":
		taskType = m.IO2
	case "CPU":
		taskType = m.CPU
	}
	taskTime, err := strconv.Atoi(task[4 : len(task)-1])
	if err != nil {
		panic(err)
	}

	return m.Task{ResouceType: taskType, TotalTime: taskTime}
}

func ParseProcess(id int, line string, logger *slog.Logger) *m.Process {
	line = strings.TrimSpace(line)
	tasks := strings.Split(line, ";")
	if tasks[len(tasks)-1] == "" {
		tasks = tasks[:len(tasks)-1]
	}
	slog.Debug(fmt.Sprintf("Tasks: %v\n", tasks))

	var parsedTasks = make([]m.Task, len(tasks))
	for i, task := range tasks {
		parsedTasks[i] = ParseTask(task)
	}
	process := m.NewProcess(id, calcArrivalTime(id), parsedTasks, logger)
	return process
}

func ParseProcesses(r io.Reader, logger *slog.Logger) []*m.Process {
	scanner := bufio.NewScanner(r)
	var processes = make([]*m.Process, 0)
	var i int
	for scanner.Scan() {
		process := ParseProcess(i, scanner.Text(), logger)
		i++
		processes = append(processes, process)
	}
	return processes
}

func snapshotState(w io.Writer, row string) {
	fmt.Fprintf(w, "%s\n", row)
}

func printProcsStats(w io.Writer, procs []*m.Process) {
	fmt.Fprintf(w, "Process\tEntrance\tService\tWaiting\tStartTime\tEndTime\tTurnaround\n")
	for _, proc := range procs {
		stats := proc.GetStats()
		fmt.Fprintf(w, "%d\t%d\t%d\t%d\t%d\t%d\t%d\n", stats.ProcId, stats.EntranceTime, stats.ServiceTime, stats.ReadyOrBlockedTime, stats.StartTime, stats.ExitTime, stats.TurnaroundTime)
	}
}

func getEvictor(schedAlgo string, procQueue *m.ProcQueue, cpuCount int) m.Evictor {
	switch schedAlgo {
	case "fcfs", "spn", "hrrn":
		return m.NewNonPreemptive()
	case "rr1":
		return m.NewRoundRobinEvictor(1)
	case "rr4":
		return m.NewRoundRobinEvictor(4)
	case "srt":
		return m.NewSRTEvictor(procQueue, cpuCount)
	default:
		panic(fmt.Sprintf("Unknown scheduling algorithm %s", schedAlgo))
	}
}

func getSelection(schedAlgo string) m.SelectionFunction {
	switch schedAlgo {
	case "fcfs":
		return m.NewSelectionFIFO()
	case "rr1":
		return m.NewSelectionFIFO()
	case "rr4":
		return m.NewSelectionFIFO()
	case "spn":
		return m.NewSelectionSPN()
	case "srt":
		return m.NewSelectionSRT()
	case "hrrn":
		return m.NewSelectionHRRN()
	default:
		panic(fmt.Sprintf("Unknown scheduling algorithm %s", schedAlgo))
	}
}

func main() {
	flag.Parse()
	var input io.Reader

	if *inputFile != "" {
		f, err := os.Open(*inputFile)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		input = f
	} else {
		input = os.Stdin
	}

	var output io.Writer

	output, err := os.Create(*outputFile)
	if err != nil {
		panic(err)
	}

	defer output.(*os.File).Close()
	snapshotFunc := func(row string) {
		snapshotState(output, row)
	}

	clock := &m.Clock{CurrentTick: 0}

	defaultHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(log.NewTickLoggerHandler(defaultHandler, clock))
	processes := ParseProcesses(input, logger)

	// IO is always fcfs
	fcfs := m.NewNonPreemptive()

	fifoSelection := m.NewSelectionFIFO()

	cpuProcQueue := m.NewProcQueue("CPUs", clock)
	evictor := getEvictor(*schedAlgo, cpuProcQueue, *cpuCount)
	selectionFunc := getSelection(*schedAlgo)

	io1ProcQueue := m.NewProcQueue("IO1", clock)
	io2ProcQueue := m.NewProcQueue("IO2", clock)

	io1Scheduler := m.NewSchedulerWrapper("IO1", io2ProcQueue, fifoSelection, fcfs, m.NewResource("IO1", m.IO1), clock, logger)
	io2Scheduler := m.NewSchedulerWrapper("IO2", io1ProcQueue, fifoSelection, fcfs, m.NewResource("IO2", m.IO2), clock, logger)
	cpuScheduler := m.NewSchedulerWrapper("CPUs", cpuProcQueue, selectionFunc, evictor, m.NewCpuPool(*cpuCount), clock, logger)

	// Run scheduler
	machine := m.NewMachine(cpuScheduler, io1Scheduler, io2Scheduler, clock, logger, snapshotFunc, *cpuCount)

	machine.Run(processes)

	procStatsFile, err := os.Create(*procStatsFile)
	if err != nil {
		panic(err)
	}

	defer procStatsFile.Close()
	printProcsStats(procStatsFile, processes)
}
