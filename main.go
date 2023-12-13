package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var (
	cpuCount = flag.Int("cpus", 4, "Number of CPUs")
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

func ParseProcess(id int, line string) Process {
	line = strings.TrimSpace(line)
	tasks := strings.Split(line, ";")
  if tasks[len(tasks)-1] == "" {
    tasks = tasks[:len(tasks)-1]
  }
  fmt.Printf("Tasks: %v\n", tasks)
	process := Process{id: id, arrivalTime: calcArrivalTime(id), state: READY, currentTaskIndex: 0, tasks: make([]Task, len(tasks))}
	for i, task := range tasks {
		process.tasks[i] = ParseTask(task)
	}
	return process
}

// reads all lines until EOF
func ParseProcesses(r io.Reader) []Process {
  scanner := bufio.NewScanner(r)
	var processes []Process
  var i int
  for scanner.Scan() {
    i++
		process := ParseProcess(i, scanner.Text())
		processes = append(processes, process)
  }
	return processes
}

func main() {
	flag.Parse()
	fmt.Println("FCFS Scheduler")

	processes := ParseProcesses(os.Stdin)

	clock := &Clock{0}
	io1Scheduler := NewFCFS("IO1", NewResource("IO1", IO1), clock)
	io2Scheduler := NewFCFS("IO2", NewResource("IO2", IO2), clock)
	cpuScheduler := NewFCFS("CPUs", NewCpuPool(*cpuCount), clock)
	// Run scheduler
	machine := NewMachine(cpuScheduler, io1Scheduler, io2Scheduler, clock)

	machine.Run(processes)
}
