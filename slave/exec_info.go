package slave

import (
	"github.com/yddeng/dutil/io"
	"github.com/yddeng/pmp/core"
	"github.com/yddeng/pmp/util"
	"path"
	"syscall"
)

const (
	dataFile = "execInfo.json"
)

var (
	execId    = int32(0)
	execInfos map[int32]*execInfo
	filename  string
)

func getExecId() int32 {
	execId++
	return execId
}

func loadExecInfo() {
	execInfos = map[int32]*execInfo{}
	filename = path.Join(core.DataPath, dataFile)
	delInfos := []int32{}
	if err := io.DecodeJsonFromFile(&execInfos, filename); err == nil {
		for id, info := range execInfos {
			if id > execId {
				execId = id
			}
			if !info.isAlive() {
				delInfos = append(delInfos, id)
			}
		}
	}

	for _, id := range delInfos {
		delete(execInfos, id)
	}
	writeExecInfo()
}

func writeExecInfo() {
	if err := io.EncodeJsonToFile(execInfos, filename); err != nil {
		util.Logger().Errorf(err.Error())
	}
}

type execInfo struct {
	ExecID  int32  `json:"exec_id,omitempty"`
	Pid     int64  `json:"pid,omitempty"`
	Name    string `json:"name,omitempty"`
	Command string `json:"command,omitempty"`
	IsGuard bool   `json:"is_guard,omitempty"`
}

func (this *execInfo) isAlive() bool {
	if err := syscall.Kill(int(this.Pid), 0); err == nil {
		return true
	}
	return false
}

func addExecInfo(execId int32, pid int64, name, command string) {
	info := &execInfo{
		ExecID:  execId,
		Name:    name,
		Pid:     pid,
		Command: command,
	}

	execInfos[execId] = info
	writeExecInfo()
}

func delExecInfo(execId int32) {
	delete(execInfos, execId)
	writeExecInfo()
}
