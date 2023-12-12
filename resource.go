package main

import "fmt"

type ResourceState int
const (
  FREE ResourceState = iota
  BUSY
)

type ResourceType int
const (
  CPU ResourceType = iota
  IO
)

type Resourcer interface {
  GetFree() (*Resource, error)
  AssignToFree(p *Process) error
}

type Resource struct {
  state ResourceState
  currentProc *Process
  ProcRunningTime int
}

type MultiCoreCpu struct {
  cpus []Resource
}

func (resource *Resource) GetFree() (*Resource, error) {
  if resource.state == BUSY {
    return nil, fmt.Errorf("Resource is busy")
  }
  return resource, nil
}

func (resource *Resource) AssignToFree(p *Process) error {
  if resource.state == BUSY {
    return fmt.Errorf("Resource is busy")
  }
  resource.state = BUSY
  resource.currentProc = p
  return nil
}

func (r *Resource) Tick() {
  if r.state == BUSY {
    r.ProcRunningTime++
  }
}

func (cpu *MultiCoreCpu) GetFree() (*Resource, error) {
  for _, res := range cpu.cpus {
    if res.state == FREE {
      return &res, nil
    }
  }
  return nil, fmt.Errorf("No available cpus")
}

func (cpu *MultiCoreCpu) AssignToFree(p *Process) error {
  for _, res := range cpu.cpus {
    if res.state == FREE {
      return res.AssignToFree(p)
    }
  }
  return fmt.Errorf("No available cpus")
}

func (cpu *MultiCoreCpu) Tick() {
  for _, res := range cpu.cpus {
    res.Tick()
  }
}
