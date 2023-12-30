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
)

var (
	cpuCount = flag.Int("cpus", 4, "Number of CPUs")
  inputFile = flag.String("input", "", "Input file")
  outputFile = flag.String("output", "result.txt", "Output file")
  procStatsFile = flag.String("procStats", "procStats.txt", "Process stats file")
)

// input format, each process starts from new line. Tasks separated by semi-colon
// CPU(12);IO2(3);CPU(4);IO1(5);CPU(2)
// CPU(4);IO1(20);CPU(10);IO1(5);CPU(2);IO2(15)

func calcArrivalTime(procId int) int {
	return procId * 2
}

func ParseTask(task string) Task {
	task = strings.TrimSpace(task)
	var taskType ResourceType
	taskTypeStr := task[:3]
	switch taskTypeStr {
	case "IO1":
		taskType = IO1
	case "IO2":
		taskType = IO2
	case "CPU":
		taskType = CPU
	}
	taskTime, err := strconv.Atoi(task[4 : len(task)-1])
	if err != nil {
		panic(err)
	}

	return Task{ResouceType: taskType, totalTime: taskTime}
}

func ParseProcess(id int, line string, logger *slog.Logger) *Process {
	line = strings.TrimSpace(line)
	tasks := strings.Split(line, ";")
  if tasks[len(tasks)-1] == "" {
    tasks = tasks[:len(tasks)-1]
  }
  slog.Debug(fmt.Sprintf("Tasks: %v\n", tasks))
	process := NewProcess(id, calcArrivalTime(id), make([]Task, len(tasks)), logger)

	for i, task := range tasks {
		process.tasks[i] = ParseTask(task)
	}
	return process
}

// reads all lines until EOF
func ParseProcesses(r io.Reader, logger *slog.Logger) []*Process {
  scanner := bufio.NewScanner(r)
	var processes = make([]*Process, 0)
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

func printProcsStats(w io.Writer, procs []*Process) {
  fmt.Fprintf(w, "Process\tEntrance\tService\tWaiting\tStartTime\tEndTime\tTurnaround\n")
  for _, proc := range procs {
    stats := proc.GetStats()
    fmt.Fprintf(w, "%d\t%d\t%d\t%d\t%d\t%d\t%d\n", stats.ProcId, stats.EntranceTime, stats.ServiceTime, stats.ReadyOrBlockedTime, stats.StartTime, stats.ExitTime, stats.TurnaroundTime)
  }
}

func main() {
	flag.Parse()
  var input io.Reader;

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

  var output io.Writer;

  output, err := os.Create(*outputFile)
  if err != nil {
    panic(err)
  }

  defer output.(*os.File).Close()
  snapshotFunc := func(row string) {
    snapshotState(output, row)
  }


	clock := &Clock{0}

  defaultHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
  logger := slog.New(NewTickLoggerHandler(defaultHandler, clock))
	processes := ParseProcesses(input, logger)

	io1Scheduler := NewFCFS("IO1", NewResource("IO1", IO1), clock, logger)
	io2Scheduler := NewFCFS("IO2", NewResource("IO2", IO2), clock, logger)
	cpuScheduler := NewFCFS("CPUs", NewCpuPool(*cpuCount), clock, logger)
	// Run scheduler
	machine := NewMachine(cpuScheduler, io1Scheduler, io2Scheduler, clock, logger, snapshotFunc)

	machine.Run(processes)

  procStatsFile, err := os.Create(*procStatsFile)
  if err != nil {
    panic(err)
  }

  defer procStatsFile.Close()
  printProcsStats(procStatsFile, processes)
}
