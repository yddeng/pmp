package slave

import (
	"fmt"
	"github.com/yddeng/dnet"
	"github.com/yddeng/dnet/drpc"
	"github.com/yddeng/pmp/core"
	"github.com/yddeng/pmp/net"
	"github.com/yddeng/pmp/protocol"
	"github.com/yddeng/pmp/util"
	"io"
	"os"
	"os/exec"
	"path"
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
	pid, err := strconv.Atoi(str)
	if nil != err {
		util.Logger().Errorln("strconv.Atoi pid error:", string(out), err)
		replyer.Reply(&protocol.StartResp{Msg: "parseInt pid error"}, nil)
		return
	}

	addExecInfo(itemID, pid)
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

var writeFiles = map[string]*writeFile{}

type writeFile struct {
	f    *os.File
	name string
}

func onSyncFile(session dnet.Session, msg *net.Message) {
	req := msg.GetData().(*protocol.File)
	if !msg.Next {
		// 表示所有数据已经发送完毕
		util.Logger().Infof("sync all file end\n")
		return
	}
	filename := req.GetFileName()

	wf, ok := writeFiles[filename]
	if !ok {
		wf = &writeFile{name: filename}
		_ = os.MkdirAll(path.Dir(filename), os.ModePerm)
		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
		if err != nil {
			panic(err)
		}
		wf.f = f
		writeFiles[filename] = wf
	}

	data := req.GetB()
	length := int(req.GetLength())
	total := 0
	for total < length {
		n, err := wf.f.Write(data[total:])
		if err != nil {
			if err != io.ErrShortWrite {
				util.Logger().Errorf(err.Error())
				break
			}
		} else {
			total += n
		}
	}

	if !req.GetNext() {
		wf.f.Close()
		delete(writeFiles, filename)
		util.Logger().Infof("sync file %s write end\n", filename)
	}

}
