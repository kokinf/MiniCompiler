package codegen

import (
	"fmt"
)

type RegisterAllocator struct {
	AvailableRegs []Register
	InUse         map[Register]bool
	SpillCount    int
}

func NewRegisterAllocator() *RegisterAllocator {
	return &RegisterAllocator{
		AvailableRegs: []Register{RAX, RCX, RDX, R8, R9, R10, R11},
		InUse:         make(map[Register]bool),
		SpillCount:    0,
	}
}

func (ra *RegisterAllocator) Allocate() (Register, error) {
	for _, reg := range ra.AvailableRegs {
		if !ra.InUse[reg] {
			ra.InUse[reg] = true
			return reg, nil
		}
	}
	return "", fmt.Errorf("no registers available")
}

func (ra *RegisterAllocator) Free(reg Register) {
	delete(ra.InUse, reg)
}

func (ra *RegisterAllocator) AllocateSpecific(reg Register) error {
	if ra.InUse[reg] {
		return fmt.Errorf("register %s already in use", reg)
	}
	ra.InUse[reg] = true
	return nil
}

func (ra *RegisterAllocator) GetSpillSlot() int {
	ra.SpillCount++
	return -(ra.SpillCount * 8)
}
