package slave

import (
	"fmt"
	"github.com/yddeng/dnet/drpc"
	"github.com/yddeng/pmp/core"
	"github.com/yddeng/pmp/protocol"
	"github.com/yddeng/pmp/util"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func onCmdStart(replyer *drpc.Replyer, req interface{}) {
	msg := req.(*protocol.StartReq)
	util.Logger().Infof("onCmdStart %v\n", msg)

	shell := fmt.Sprintf("nohup %s %s %s > /dev/null 2> /dev/null & echo $!", msg.GetName(), msg.GetCommand(), core.OpArg)
	cmd := exec.Command("sh", "-c", shell)
	out, err := cmd.Output()
	if err != nil {
		replyer.Reply(&protocol.StartResp{Msg: err.Error()}, nil)
		util.Logger().Errorf(err.Error())
		return
	}

	// 进程pid
	str := strings.Split(string(out), "\n")[0]
	pid, err := strconv.ParseInt(str, 10, 64)
	if nil != err {
		util.Logger().Errorln("parseInt pid error:", string(out), err)
		replyer.Reply(&protocol.StartResp{Msg: "parseInt pid error"}, nil)
		return
	}

	addExecInfo(getExecId(), pid, msg.GetName(), msg.GetCommand())
	util.Logger().Infof("start ok, pid %d", pid)
	replyer.Reply(&protocol.StartResp{}, nil)
}

func onCmdStop(replyer *drpc.Replyer, req interface{}) {
	msg := req.(*protocol.StopReq)
	util.Logger().Infof("onCmdStop %v\n", msg)

	execId := msg.GetExecId()
	p, ok := execInfos[execId]
	if ok {
		err := syscall.Kill(int(p.Pid), syscall.SIGTERM)
		if err != nil {
			replyer.Reply(&protocol.StopResp{Msg: err.Error()}, nil)
			util.Logger().Errorf(err.Error())
			return
		}
		delExecInfo(execId)
	}

	util.Logger().Infof("stop ok, execId %d", execId)
	replyer.Reply(&protocol.StopResp{}, nil)
}

func onCmdKill(replyer *drpc.Replyer, req interface{}) {
	msg := req.(*protocol.KillReq)
	util.Logger().Infof("onCmdKill %v\n", msg)

	execId := msg.GetExecId()
	p, ok := execInfos[execId]
	if ok {
		err := syscall.Kill(int(p.Pid), syscall.SIGKILL)
		if err != nil {
			replyer.Reply(&protocol.KillResp{Msg: err.Error()}, nil)
			util.Logger().Errorf(err.Error())
			return
		}
		delExecInfo(execId)
	}

	util.Logger().Infof("kill ok, execId %d", execId)
	replyer.Reply(&protocol.KillResp{}, nil)
}

func onCmdSignal(replyer *drpc.Replyer, req interface{}) {
	msg := req.(*protocol.SignalReq)
	util.Logger().Infof("onCmdSignal %v\n", msg)

	execId := msg.GetExecId()
	p, ok := execInfos[execId]
	if ok {
		err := syscall.Kill(int(p.Pid), syscall.Signal(int(msg.GetSignal())))
		if err != nil {
			replyer.Reply(&protocol.SignalResp{Msg: err.Error()}, nil)
			util.Logger().Errorf(err.Error())
			return
		}
	}

	util.Logger().Infof("signal ok, execId %d", execId)
	replyer.Reply(&protocol.SignalResp{}, nil)
}
