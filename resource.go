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

type Holdable interface {
  Hold(index int)
  Release(index int)
}

type Resource struct {
  state ResourceState
  currentProc *Process
}

type ResourceSet []Resource

func (rs ResourceSet) Hold(index int) {
  rs[index].state = BUSY
}

func (rs ResourceSet) Release(index int) {
  rs[index].state = FREE
}

func (rs ResourceSet) GetFree() (*Resource, error) {
  for _, res := range rs {
    if res.state == FREE {
      return &res, nil
    }
  }
  return nil, fmt.Errorf("No available resources")
}

func AssignToResource(resource *Resource, process *Process) {
  if resource.state == BUSY {
    panic("Resource is busy")
  }
  resource.state = BUSY
  process.state = RUNNING
  resource.currentProc = process
}

