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

	itemID := msg.GetItemID()
	_, ok := execInfos[itemID]
	if ok {
		replyer.Reply(&protocol.StartResp{Msg: "itemID is started"}, nil)
		return
	}

	shell := fmt.Sprintf("nohup %s %s > /dev/null 2> /dev/null & echo $!", msg.GetArgs(), core.OpArg)
	util.Logger().Debugln(itemID, shell)
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

	addExecInfo(itemID, pid, msg.GetArgs())
	util.Logger().Infof("start ok,itemID %d pid %d", itemID, pid)
	replyer.Reply(&protocol.StartResp{}, nil)
}

func onCmdSignal(replyer *drpc.Replyer, req interface{}) {
	msg := req.(*protocol.SignalReq)
	util.Logger().Infof("onCmdSignal %v\n", msg)

	itemID := msg.GetItemID()
	p, ok := execInfos[itemID]
	if !ok {
		replyer.Reply(&protocol.SignalResp{Msg: "itemID not exist"}, nil)
		return
	}

	var err error
	switch msg.GetSignal() {
	case protocol.Signal_term:
		err = syscall.Kill(int(p.Pid), syscall.SIGTERM)
		delExecInfo(itemID)
	case protocol.Signal_kill:
		err = syscall.Kill(int(p.Pid), syscall.SIGKILL)
		delExecInfo(itemID)
	case protocol.Signal_user1:
		err = syscall.Kill(int(p.Pid), syscall.SIGUSR1)
	case protocol.Signal_user2:
		err = syscall.Kill(int(p.Pid), syscall.SIGUSR2)
	}
	if err != nil {
		replyer.Reply(&protocol.SignalResp{Msg: err.Error()}, nil)
		util.Logger().Errorf(err.Error())
		return
	}

	util.Logger().Infof("signal ok, itemID %d", itemID)
	replyer.Reply(&protocol.SignalResp{}, nil)
}
