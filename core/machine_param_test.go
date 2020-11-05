package core

import (
	"fmt"
	"testing"
	"time"
)

func TestMachineParam_CPU(t *testing.T) {
	cpu := &MachineParam{}

	//cpu.CPU(false)
	for {
		time.Sleep(time.Second)
		f, _ := cpu.CPU(false)
		if len(f) > 0 {
			fmt.Println(0, f[0])
		}
	}

	time.Sleep(time.Second * 5)
}
