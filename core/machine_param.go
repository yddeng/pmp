package core

import (
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/yddeng/pmp/util"
	"math"
)

type MachineParam struct {
	oldCpus *cpu.TimesStat
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

// cpuCount , usedPercent
func (this *MachineParam) CPU() (int, float64) {
	if this.oldCpus == nil {
		stat, _ := cpu.Times(false)
		this.oldCpus = &stat[0]
	}

	stat, _ := cpu.Times(false)
	oldCpu := *this.oldCpus
	this.oldCpus = &stat[0]
	nowCpu := stat[0]

	count, _ := cpu.Counts(false)
	return count, calculateBusy(oldCpu, nowCpu)
}

// total, used , usedPercent
func (this *MachineParam) Mem() (uint64, uint64, float64) {
	mMem, _ := mem.VirtualMemory()
	return mMem.Total, mMem.Used, mMem.UsedPercent

}

func (this *MachineParam) MemFormat() (string, string, float64) {
	mMem, _ := mem.VirtualMemory()
	total := util.B2String(mMem.Total, 1024)
	used := util.B2String(mMem.Used, 1024)
	return total, used, mMem.UsedPercent
}

// total, used , usedPercent
func (this *MachineParam) Disk() (uint64, uint64, float64) {
	mDisk, _ := disk.Usage("/")
	return mDisk.Total, mDisk.Total - mDisk.Free, float64(mDisk.Total-mDisk.Free) * 100 / float64(mDisk.Total)
}

func (this *MachineParam) DiskFormat() (string, string, float64) {
	mDisk, _ := disk.Usage("/")
	total := util.B2String(mDisk.Total, 1000)
	used := util.B2String(mDisk.Total-mDisk.Free, 1000)
	return total, used, float64(mDisk.Total-mDisk.Free) * 100 / float64(mDisk.Total)
}
