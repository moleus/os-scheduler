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
	schedAlgo     = flag.String("algo", "fcfs", "Scheduling algorithm (default: fcfs). Possible values: fcfs, rr1, rr2, spn, srt, hrrn, rr")
  roundRobinQuantum = flag.Int("quantum", 4, "Round robin quantum (default: 4)")
  arrivalInterval = flag.Int("interval", 2, "Proc arrival interval (default: 2)")
  logLevel = flag.String("log", "debug", "Log level (default: debug)")
)

func calcArrivalTime(procId int) int {
	return procId * *arrivalInterval
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

func ParseProcess(id int, line string, logger *slog.Logger, clock log.GlobalTimer) *m.Process {
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
	process := m.NewProcess(id, calcArrivalTime(id), parsedTasks, logger, clock)
	return process
}

func ParseProcesses(r io.Reader, logger *slog.Logger, clock log.GlobalTimer) []*m.Process {
	scanner := bufio.NewScanner(r)
	var processes = make([]*m.Process, 0)
	var i int
	for scanner.Scan() {
		process := ParseProcess(i, scanner.Text(), logger, clock)
		i++
		processes = append(processes, process)
	}
	return processes
}

func snapshotState(w io.Writer, row string) {
	fmt.Fprintf(w, "%s\n", row)
}

func printProcsStats(w io.Writer, procs []*m.Process) {
	fmt.Fprintf(w, "Process\tArrival\tService\tWaiting\tFinish time\tTurnaround (Tr)\tTr/Ts\n")
	for _, proc := range procs {
		stats := proc.GetStats()
		normalizedTurnaround := float64(stats.TurnaroundTime) / float64(stats.ServiceTime)
		fmt.Fprintf(w, "%d\t%d\t%d\t%d\t%d\t%d\t%f\n", stats.ProcId, stats.EntranceTime, stats.ServiceTime, stats.ReadyOrBlockedTime, stats.ExitTime, stats.TurnaroundTime, normalizedTurnaround)
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
  case "rr":
    return m.NewRoundRobinEvictor(*roundRobinQuantum)
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
	case "rr1", "rr4", "rr":
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

func parseLogLevel(level string) slog.Level {
  switch level {
  case "debug":
    return slog.LevelDebug
  case "info":
    return slog.LevelInfo
  case "warn":
    return slog.LevelWarn
  case "error":
    return slog.LevelError
  default:
    panic(fmt.Sprintf("Unknown log level %s", level))
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

  logLevel := parseLogLevel(*logLevel)
	defaultHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	logger := slog.New(log.NewTickLoggerHandler(defaultHandler, clock))
	processes := ParseProcesses(input, logger, clock)

  logger.Info(fmt.Sprintf("Running with %d CPUs", *cpuCount))
  logger.Info(fmt.Sprintf("Total processes: %d", len(processes)))

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
