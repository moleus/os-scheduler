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

func (cpu *MultiCoreCpu) GetFree() (*Resource, error) {
  for _, res := range cpu.cpus {
    if res.state == FREE {
      return &res, nil
    }
  }
  return nil, fmt.Errorf("No available cpus")
}

