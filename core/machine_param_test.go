package core

import (
	"fmt"
	"testing"
	"time"
)

func TestMachineParam_CPU(t *testing.T) {
	sys := &MachineParam{}

	for {
		time.Sleep(time.Second)
		n, f := sys.CPU()
		fmt.Println(n, f)
		fmt.Println(sys.MemFormat())
		fmt.Println(sys.DiskFormat())
	}

}
