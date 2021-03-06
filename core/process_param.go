package core

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

const (
	paramFive   = "12345"
	paramTen    = paramFive + paramFive
	paramTwenty = paramTen + paramTen
	paramFifty  = paramTwenty + paramTwenty + paramTen
)

// 获取cpu使用情况。 返回%前面值, 0.2%  return 0.2
func ProcessCPUUsed(pid int) (float64, error) {
	output, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "pcpu=12345").Output()
	if err != nil {
		return 0, err
	}

	linesOfProcStrings := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(linesOfProcStrings) < 2 {
		return 0, fmt.Errorf("linesOfProcStrings failed %v ", linesOfProcStrings)
	}

	line := linesOfProcStrings[1]
	n, err := strconv.ParseFloat(strings.TrimSpace(line[0:5]), 64)
	if err != nil {
		return 0, err
	}

	return n, nil
}

// 获取内存使用情况。 返回%前面值, 0.2%  return 0.2
func ProcessMemUsed(pid int) (float64, error) {
	output, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "pmem=12345").Output()
	if err != nil {
		return 0, err
	}

	linesOfProcStrings := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(linesOfProcStrings) < 2 {
		return 0, fmt.Errorf("linesOfProcStrings failed %v ", linesOfProcStrings)
	}

	line := linesOfProcStrings[1]
	n, err := strconv.ParseFloat(strings.TrimSpace(line[0:5]), 64)
	if err != nil {
		return 0, err
	}

	return n, nil

}

func ProcessMemCpuUsed(pid int) (float64, float64, error) {
	output, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "pmem=12345,pcpu=12345").Output()
	if err != nil {
		return 0, 0, err
	}

	linesOfProcStrings := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(linesOfProcStrings) < 2 {
		return 0, 0, fmt.Errorf("linesOfProcStrings failed %v ", linesOfProcStrings)
	}

	line := linesOfProcStrings[1]
	mem, err := strconv.ParseFloat(strings.TrimSpace(line[0:5]), 64)
	if err != nil {
		return 0, 0, err
	}
	cpu, err := strconv.ParseFloat(strings.TrimSpace(line[6:11]), 64)
	if err != nil {
		return 0, 0, err
	}

	return mem, cpu, nil
}
