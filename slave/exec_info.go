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
	execInfos map[int32]*execInfo
	filename  string
)

func loadExecInfo() {
	execInfos = map[int32]*execInfo{}
	filename = path.Join(core.DataPath, dataFile)
	delInfos := []int32{}
	if err := io.DecodeJsonFromFile(&execInfos, filename); err == nil {
		for id, info := range execInfos {
			isAlive := info.isAlive()
			util.Logger().Infof("loadExecInfo %v isAlive:%v", info, isAlive)
			if !isAlive {
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
	ItemID int32 `json:"item_id,omitempty"`
	Pid    int   `json:"pid,omitempty"`
}

func (this *execInfo) isAlive() bool {
	if err := syscall.Kill(this.Pid, 0); err == nil {
		return true
	}
	return false
}

func addExecInfo(itemID int32, pid int) *execInfo {
	info := &execInfo{
		ItemID: itemID,
		Pid:    pid,
	}

	execInfos[itemID] = info
	writeExecInfo()
	return info
}

func delExecInfo(execId int32) {
	delete(execInfos, execId)
	writeExecInfo()
}
