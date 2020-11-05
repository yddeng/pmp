package core

import (
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"math"
)

const (
	paramFive   = "12345"
	paramTen    = paramFive + paramFive
	paramTwenty = paramTen + paramTen
	paramFifty  = paramTwenty + paramTwenty + paramTen
)

type MachineParam struct {
	oldCpus   []cpu.TimesStat
	oldPercpu bool
}

func getAllBusy(t cpu.TimesStat) (float64, float64) {
	busy := t.User + t.System + t.Nice + t.Iowait + t.Irq +
		t.Softirq + t.Steal
	return busy + t.Idle, busy
}

func calculateBusy(t1, t2 cpu.TimesStat) float64 {
	t1All, t1Busy := getAllBusy(t1)
	t2All, t2Busy := getAllBusy(t2)

	if t2Busy <= t1Busy {
		return 0
	}
	if t2All <= t1All {
		return 100
	}
	return math.Min(100, math.Max(0, (t2Busy-t1Busy)/(t2All-t1All)*100))
}

func calculateAllBusy(t1, t2 []cpu.TimesStat) ([]float64, error) {
	// Make sure the CPU measurements have the same length.
	if len(t1) != len(t2) {
		return nil, fmt.Errorf(
			"received two CPU counts: %d != %d",
			len(t1), len(t2),
		)
	}

	ret := make([]float64, len(t1))
	for i, t := range t2 {
		ret[i] = calculateBusy(t1[i], t)
	}
	return ret, nil
}

func (this *MachineParam) CPU(percpu bool) ([]float64, error) {
	var err error
	if this.oldPercpu != percpu || len(this.oldCpus) == 0 {
		this.oldCpus, err = cpu.Times(percpu)
		this.oldPercpu = percpu
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	stat, err := cpu.Times(percpu)
	if err != nil {
		return nil, err
	}
	return calculateAllBusy(this.oldCpus, stat)
}

func (this *MachineParam) Mem() {

}

func (this *MachineParam) Disk() {

}

func (this *MachineParam) AllProcess() {

}
