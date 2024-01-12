package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/xuri/excelize/v2"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"

	log "github.com/Moleus/os-solver/pkg/logging"
	m "github.com/Moleus/os-solver/pkg/machine"
)

var (
	cpuCount          = flag.Int("cpus", 4, "Number of CPUs")
	inputFile         = flag.String("input", "", "Input file")
	outputFile        = flag.String("output", "result.txt", "Output file")
	procStatsFile     = flag.String("procStats", "procStats.txt", "Process stats file")
	schedAlgo         = flag.String("algo", "fcfs", "Scheduling algorithm (default: fcfs). Possible values: fcfs, rr1, rr2, spn, srt, hrrn, rr")
	roundRobinQuantum = flag.Int("quantum", 4, "Round robin quantum (default: 4)")
	arrivalInterval   = flag.Int("interval", 2, "Proc arrival interval (default: 2)")
	logLevel          = flag.String("log", "debug", "Log level (default: debug)")
	exportXlsx        = flag.String("export-xlsx", "", "Path for creating xlsx report")
)

const countOfHardcodedColors = 10

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
func snapshotStateXlsx(f *excelize.File, sheet string, tick string, cpusStateString []string, io1State string, io2State string, colors [countOfHardcodedColors]int) {
	err := f.SetCellValue(sheet, fmt.Sprintf("A%s", tick), tick)
	if err != nil {
		return
	}
	for pos, val := range cpusStateString {
		err := f.SetCellValue(sheet, fmt.Sprintf("%s%s", fmt.Sprintf("%c", 'A'+pos+1), tick), val)
		if err != nil {
			return
		}
		setStyle(f, sheet, fmt.Sprintf("%s%s", fmt.Sprintf("%c", 'A'+pos+1), tick), val, colors)
	}
	err = f.SetCellValue(sheet, fmt.Sprintf("%s%s", fmt.Sprintf("%c", 'A'+*cpuCount+1), tick), io1State)
	if err != nil {
		return
	}
	setStyle(f, sheet, fmt.Sprintf("%s%s", fmt.Sprintf("%c", 'A'+*cpuCount+1), tick), io1State, colors)
	err = f.SetCellValue(sheet, fmt.Sprintf("%s%s", fmt.Sprintf("%c", 'A'+*cpuCount+2), tick), io2State)
	if err != nil {
		return
	}
	setStyle(f, sheet, fmt.Sprintf("%s%s", fmt.Sprintf("%c", 'A'+*cpuCount+2), tick), io2State, colors)
}
func setStyle(f *excelize.File, spreed string, cell string, val string, colors [countOfHardcodedColors]int) {
	if val == "-" {
		return
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		panic(err)
	}
	if i > countOfHardcodedColors {
		return
	}
	err = f.SetCellStyle(spreed, cell, cell, colors[i])
	if err != nil {
		panic(err)
	}
}
func getF() *excelize.File {
	var f *excelize.File
	if _, err := os.Stat(*exportXlsx); errors.Is(err, os.ErrNotExist) {
		f = excelize.NewFile()

	} else {
		f, err = excelize.OpenFile(*exportXlsx)
		if err != nil {
			panic(err)
		}
		err = f.Close()
		if err != nil {
			panic(err)
		}
	}
	index, err := f.NewSheet(*schedAlgo)
	if err != nil {
		panic(err)
	}
	f.SetActiveSheet(index)
	err = f.DeleteSheet("Sheet1")
	if err != nil {
		return nil
	}
	return f
}
func printRow(f *excelize.File, sheet string, offset int, row int, values []string) {
	for pos, val := range values {
		err := f.SetCellValue(sheet, fmt.Sprintf("%s%d", fmt.Sprintf("%c", 'A'+offset+pos), row), val)
		if err != nil {
			return
		}
	}
}
func generateStyles(f *excelize.File) [countOfHardcodedColors]int {
	colors := [countOfHardcodedColors]string{"E0EBF5", "#93e476", "#efb2b9", "#6a74eb", "#f0b1e4", "#c1b1f0", "#ead669", "#ebaa6a", "#eb836a", "#6aeb71"}
	var styles [countOfHardcodedColors]int
	for i := 0; i < countOfHardcodedColors; i++ {
		style, err := f.NewStyle(&excelize.Style{
			Fill: excelize.Fill{Type: "pattern", Color: []string{colors[i]}, Pattern: 1},
		})
		if err != nil {
			fmt.Println(err)
		}
		styles[i] = style
	}
	return styles
}

func printProcsStats(w io.Writer, procs []*m.Process) {
	fmt.Fprintf(w, "Process\tArrival\tService\tWaiting\tFinish time\tTurnaround (Tr)\tTr/Ts\n")
	for _, proc := range procs {
		stats := proc.GetStats()
		normalizedTurnaround := float64(stats.TurnaroundTime) / float64(stats.ServiceTime)
		fmt.Fprintf(w, "%d\t%d\t%d\t%d\t%d\t%d\t%f\n", stats.ProcId+1, stats.EntranceTime, stats.ServiceTime, stats.ReadyOrBlockedTime, stats.ExitTime, stats.TurnaroundTime, normalizedTurnaround)
	}
}

func getScheduler(schedAlgo string, procQueue *m.ProcQueue, cpuCount int) (m.Evictor, m.SelectionFunction) {
	switch schedAlgo {
	case "fcfs":
		return m.NewNonPreemptive(), m.NewSelectionFIFO()
	case "rr1":
		return m.NewRoundRobinEvictor(1), m.NewSelectionFIFO()
	case "rr4":
		return m.NewRoundRobinEvictor(4), m.NewSelectionFIFO()
	case "rr":
		return m.NewRoundRobinEvictor(*roundRobinQuantum), m.NewSelectionFIFO()
	case "spn":
		return m.NewNonPreemptive(), m.NewSelectionSPN()
	case "srt":
		srt := m.NewSchedulerSRT(procQueue, cpuCount)
		return srt, srt
	case "hrrn":
		return m.NewNonPreemptive(), m.NewSelectionHRRN()
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
	snapshotFunc := func(tick string, cpu []string, io1 string, io2 string) {
		snapshotState(output, fmt.Sprintf("%3s %s %s %s", tick, strings.Join(cpu, " "), io1, io2))
	}
	var f *excelize.File
	if *exportXlsx != "" {
		f = getF()
		colors := generateStyles(f)
		snapshotFunc = func(tick string, cpu []string, io1 string, io2 string) {
			snapshotState(output, fmt.Sprintf("%3s %s %s %s", tick, strings.Join(cpu, " "), io1, io2))
			snapshotStateXlsx(f, *schedAlgo, tick, cpu, io1, io2, colors)
		}
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
	evictor, selectionFunc := getScheduler(*schedAlgo, cpuProcQueue, *cpuCount)

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
	if *exportXlsx != "" {
		if err := f.SaveAs(*exportXlsx); err != nil {
			fmt.Println(err)
		}
	}
}
