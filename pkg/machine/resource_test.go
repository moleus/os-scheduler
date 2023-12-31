package machine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssignToFreeCpuPool(t *testing.T) {
	cpuPool := NewCpuPool(2)
	p := &Process{id: 0}
	cpu, err := cpuPool.GetFree()
	assert.Nil(t, err)
	err = cpu.AssignToFree(p)
	assert.Nil(t, err)
	assert.Equal(t, BUSY, cpu.state)
	assert.Equal(t, BUSY, cpuPool.cpus[0].state)
	assert.Equal(t, p, cpu.currentProc)

	p2 := &Process{id: 1}
	assert.Nil(t, err)
	err = cpuPool.AssignToFree(p2)
	assert.Nil(t, err)

	freeCpu, err := cpuPool.GetFree()
	assert.Nilf(t, freeCpu, "Expected no free CPU, but got one %s", freeCpu)
	assert.NotNil(t, err, "Expected an error when no CPUs, but err is nil")

	assert.Equal(t, BUSY, cpuPool.cpus[1].state)
	assert.Equal(t, p2, cpuPool.cpus[1].currentProc)

	procs := cpuPool.GetProcs()
	assert.Equal(t, 2, len(procs))
}
