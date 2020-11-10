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
	"os"
	"path/filepath"
	"sync"
)

var (
	slavePtr *slave
)

type slave struct {
	mtx    sync.RWMutex
	slaves map[string]*Slave
}

func (this *slave) getAll() map[string]*Slave {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	return this.slaves
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
	ok      bool
	items   map[int32]*Item
}

type Item struct {
	itemID int32
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

func onClientClose(session dnet.Session, reason string) {
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
	_, ok := slavePtr.get(name)
	if ok {
		util.Logger().Infof("slave %s is already login", name)
		replyer.Reply(&protocol.LoginResp{Msg: "already login"}, nil)
		return
	}

	slave.name = name
	slavePtr.set(name, slave)
	slave.session.SetContext(slave)

	go func() {
		if err := getAndSyncAll(slave); err != nil {
			slave.session.Close(err.Error())
		}
	}()
	replyer.Reply(&protocol.LoginResp{}, nil)
}

func getAndSyncAll(slave *Slave) error {
	length := net.BuffSize - net.HeadSize - 200
	err := filepath.Walk(core.FileSyncPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			idx := 0
			total := len(data)
			for total > length {
				file := &protocol.File{
					FileName: path,
					B:        data[idx : idx+length],
					Next:     true,
					Length:   int32(length),
				}
				slave.SendMessage(file, true)
				idx += length
				total -= length
			}

			file := &protocol.File{
				FileName: path,
				B:        data[idx:],
				Length:   int32(total),
			}
			slave.SendMessage(file, true)

		}
		return nil
	})
	if err != nil {
		return err
	}

	slave.SendMessage(&protocol.File{})
	eventQueue.Push(func() {
		slave.ok = true
	})
	return nil
}
