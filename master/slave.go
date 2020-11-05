package master

import (
	"github.com/golang/protobuf/proto"
	"github.com/yddeng/dnet"
	"github.com/yddeng/dnet/drpc"
	"github.com/yddeng/pmp/core"
	"github.com/yddeng/pmp/net"
	"github.com/yddeng/pmp/protocol"
	"github.com/yddeng/pmp/util"
	"time"
)

var (
	slaves map[string]*Slave
)

type Slave struct {
	name    string
	session dnet.Session
	ok      bool
}

func getSlave() *Slave {
	for _, s := range slaves {
		return s
	}
	return nil
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

func onClientClose(session dnet.Session, reason string) {
	eventQueue.Push(func() {
		ctx := session.Context()
		if ctx != nil {
			slave := ctx.(*Slave)
			util.Logger().Infof("slave %s Close %s", slave.name, reason)
			delete(slaves, slave.name)
		}
	})
}

func onLogin(replyer *drpc.Replyer, req interface{}) {
	slave := replyer.Channel.(*Slave)
	msg := req.(*protocol.LoginReq)

	name := msg.GetName()
	util.Logger().Infof("slave %s is login\n", name)
	_, ok := slaves[name]
	if ok {
		util.Logger().Infof("slave %s is already login", name)
		replyer.Reply(&protocol.LoginResp{Msg: "already login"}, nil)
		return
	}

	slave.name = name
	slaves[name] = slave
	slave.session.SetContext(slave)

	go func() {
		if err := getAndSyncAll(slave); err != nil {
			slave.session.Close(err.Error())
		}
	}()
	replyer.Reply(&protocol.LoginResp{}, nil)

	go func() {
		time.Sleep(time.Second * 5)
		Start()
	}()
}
