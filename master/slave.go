package master

import (
	"github.com/golang/protobuf/proto"
	"github.com/yddeng/dnet"
	"github.com/yddeng/dnet/drpc"
	"github.com/yddeng/pmp/core"
	"github.com/yddeng/pmp/net"
	"github.com/yddeng/pmp/protocol"
	"github.com/yddeng/pmp/util"
	"io/ioutil"
	"path"
	"sync"
)

var (
	slavePtr *slave
)

type slave struct {
	mtx    sync.RWMutex
	slaves map[string]*Slave
}

func (this *slave) allDo(call func(slave2 *Slave), except string) {
	this.mtx.RLock()
	for _, s := range this.slaves {
		if except != "" && s.name != except {
			this.mtx.RUnlock()
			call(s)
			this.mtx.RLock()
		}
	}
	this.mtx.RUnlock()
}

func (this *slave) getAll() map[string]*Slave {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	return this.slaves
}

func (this *slave) getRunInfo() map[int32]*protocol.ItemInfo {
	runInfo := map[int32]*protocol.ItemInfo{}
	this.mtx.Lock()
	defer this.mtx.Unlock()
	for _, s := range this.slaves {
		for _, v := range s.Report.GetItems() {
			runInfo[v.GetItemID()] = v
		}
	}
	return runInfo
}

func (this *slave) get(key string) (*Slave, bool) {
	this.mtx.RLock()
	s, ok := this.slaves[key]
	this.mtx.RUnlock()
	return s, ok
}

func (this *slave) set(key string, s *Slave) {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	this.slaves[key] = s
}

func (this *slave) delete(key string) {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	delete(this.slaves, key)
}

type Slave struct {
	name    string
	session dnet.Session
	Report  *protocol.Report `json:"report"`
	mtx     sync.RWMutex
}

func (this *Slave) GetReport() *protocol.Report {
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	return this.Report
}

func (this *Slave) send(msg interface{}) error {
	err := this.session.Send(msg)
	if err != nil {
		util.Logger().Errorln(this.name, "send", err.Error())
	}
	return err
}

func (this *Slave) SendMessage(msg proto.Message, next ...bool) error {
	return this.send(net.NewMessage(msg, next...))
}

func (this *Slave) SendRequest(req *drpc.Request) error {
	return this.send(req)
}

func (this *Slave) SendResponse(resp *drpc.Response) error {
	return this.send(resp)
}

func (this *Slave) AsynCall(data proto.Message, callback func(interface{}, error)) error {
	return rpcClient.AsynCall(this, proto.MessageName(data), data, core.RpcTimeout, callback)
}

func (this *Slave) SyncCall(data proto.Message) (ret interface{}, err error) {
	ch := make(chan bool, 1)
	if err = this.AsynCall(data, func(i interface{}, e error) {
		ret, err = i, e
		ch <- true
	}); err == nil {
		<-ch
	}
	return
}

func onClose(session dnet.Session, reason string) {
	eventQueue.Push(func() {
		ctx := session.Context()
		if ctx != nil {
			slave := ctx.(*Slave)
			util.Logger().Infof("slave %s Close %s", slave.name, reason)
			slavePtr.delete(slave.name)
		}
	})
}

func onLogin(replyer *drpc.Replyer, req interface{}) {
	slave := replyer.Channel.(*Slave)
	msg := req.(*protocol.LoginReq)

	name := msg.GetName()
	util.Logger().Infof("slave %s is login\n", name)

	if _, ok := slavePtr.get(name); ok {
		replyer.Reply(&protocol.LoginResp{Msg: "name already login"}, nil)
		return
	}

	slave.name = name
	slavePtr.set(slave.name, slave)
	slave.session.SetContext(slave)
	replyer.Reply(&protocol.LoginResp{}, nil)

	if name != "master" {
		go func() {
			filePtr.mtx.RLock()
			if err := syncAllFile(slave, filePtr.FileInfo); err != nil {
				slave.session.Close(err.Error())
			}
			filePtr.mtx.RUnlock()
		}()
	}
}

func syncAllFile(slave *Slave, info *fileInfo) (err error) {
	if info == nil {
		return
	}
	for _, cInfo := range info.FileInfos {
		if cInfo.IsDir {
			err = syncAllFile(slave, cInfo)
		} else {
			err = syncFile(slave, path.Join(cInfo.Path, cInfo.Name))
		}
		if err != nil {
			return
		}
	}
	return
}

func syncFile(slave *Slave, filename string) error {
	length := net.BuffSize - net.HeadSize - 200
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	idx := 0
	total := len(data)
	for total > length {
		file := &protocol.File{
			FileName: filename,
			B:        data[idx : idx+length],
			Next:     true,
			Length:   int32(length),
		}
		slave.SendMessage(file, true)
		idx += length
		total -= length
	}

	file := &protocol.File{
		FileName: filename,
		B:        data[idx:],
		Length:   int32(total),
	}
	slave.SendMessage(file, true)
	return nil
}

func slaveReport(slave *Slave, msg *net.Message) {
	report := msg.GetData().(*protocol.Report)
	slave.mtx.Lock()
	slave.Report = report
	slave.mtx.Unlock()

	//for _, v := range report.GetItems() {
	//
	//}
}
