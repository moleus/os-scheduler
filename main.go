package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
  cpuCount = flag.Int("cpus", 4, "Number of CPUs")
)

func main() {
  flag.Parse()
  fmt.Println("FCFS Scheduler")

  // Get input
  reader := bufio.NewReader(os.Stdin)
  fmt.Print("Enter number of processes: ")
  numProcStr, _ := reader.ReadString('\n')
  numProc, _ := strconv.Atoi(strings.TrimSpace(numProcStr))

  // Create processes
  processes := make([]Process, numProc)
  for i := 0; i < numProc; i++ {
    processes[i].id = i
    fmt.Printf("Enter arrival time for process %d: ", i)
    arrivalTimeStr, _ := reader.ReadString('\n')
    processes[i].arrivalTime, _ = strconv.Atoi(strings.TrimSpace(arrivalTimeStr))
  }

  io1Scheduler := NewFCFS(&Resource{})
  io2Scheduler := NewFCFS(&Resource{})
  cpuScheduler := NewFCFS(NewCpuPool(*cpuCount))
  // Run scheduler
  machine := NewMachine(cpuScheduler, io1Scheduler, io2Scheduler)

  machine.Run(processes)
}
