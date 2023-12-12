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
  MustEvict(p *Process)
  GetProcs() []Process
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

func (r *Resource) GetProcs() []Process {
  if r.state == BUSY {
    return []Process{*r.currentProc}
  }
  return []Process{}
}

func (r *Resource) MustEvict(p *Process) {
  if r.state == FREE {
    panic(fmt.Sprintf("Can't evict process. Resource is free"))
  }
  if r.currentProc.id != p.id {
    panic(fmt.Sprintf("Process %d is not running on resource", p.id))
  }
  r.state = FREE
  r.currentProc = nil
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

func (cpu *MultiCoreCpu) GetProcs() []Process {
  procs := []Process{}
  for _, res := range cpu.cpus {
    procs = append(procs, res.GetProcs()...)
  }
  return procs
}

func (cpu *MultiCoreCpu) MustEvict(p *Process) {
  for _, res := range cpu.cpus {
    if res.state == BUSY && res.currentProc.id == p.id {
      res.MustEvict(p)
      return
    }
  }
  panic(fmt.Sprintf("Process %d is not running on cpu", p.id))
}
