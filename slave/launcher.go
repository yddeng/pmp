package slave

import (
	"github.com/golang/protobuf/proto"
	"github.com/yddeng/dnet"
	"github.com/yddeng/dnet/drpc"
	"github.com/yddeng/dnet/dtcp"
	"github.com/yddeng/dutil/event"
	"github.com/yddeng/pmp/core"
	"github.com/yddeng/pmp/net"
	"github.com/yddeng/pmp/net/pb"
	"github.com/yddeng/pmp/protocol"
	"github.com/yddeng/pmp/util"
	"log"
	"time"
)

const (
	state_none    = 0
	state_dailing = 1
	state_ok      = 2
)

type Launcher struct {
	state            int32
	session          dnet.Session
	name, masterAddr string
	eventQue         *event.EventQueue
	handler          map[uint16]func(session dnet.Session, msg *net.Message)
	rpcServer        *drpc.Server
	rpcClient        *drpc.Client
}

func (this *Launcher) send(msg interface{}) error {
	err := this.session.Send(msg)
	if err != nil {
		util.Logger().Errorf(err.Error())
	}
	return err
}

func (this *Launcher) SendRequest(req *drpc.Request) error {
	return this.send(req)
}

func (this *Launcher) SendResponse(resp *drpc.Response) error {
	return this.send(resp)
}

func (this *Launcher) AsynCall(data proto.Message, callback func(interface{}, error)) error {
	return this.rpcClient.AsynCall(this, proto.MessageName(data), data, core.RpcTimeout, callback)
}

func (this *Launcher) dial() {
	if this.session != nil || this.state != state_none {
		return
	}

	this.state = state_dailing

	go func() {
		for {
			session, err := dtcp.DialTCP("tcp", this.masterAddr, time.Second*5)
			if nil == err && session != nil {
				this.onConnected(session)
				return
			} else {
				time.Sleep(1 * time.Second)
			}
		}
	}()
}

func (this *Launcher) onConnected(session dnet.Session) {
	this.eventQue.Push(func() {
		this.session = session
		session.SetCodec(net.NewCodec("pmp_msg", "pmp_req", "pmp_resp"))
		session.SetCloseCallBack(func(session dnet.Session, reason string) {
			this.eventQue.Push(func() {
				this.session = nil
				this.state = state_none
				log.Printf("session closed, reason: %s\n", reason)
			})
		})

		_ = session.Start(func(data interface{}, err error) {
			if err != nil {
				session.Close(err.Error())
			} else {
				this.eventQue.Push(func() {
					switch data.(type) {
					case *drpc.Request:
						this.rpcServer.OnRPCRequest(this, data.(*drpc.Request))
					case *drpc.Response:
						this.rpcClient.OnRPCResponse(data.(*drpc.Response))
					case *net.Message:
						this.dispatchMsg(session, data.(*net.Message))
					}
				})
			}
		})

		login := &protocol.LoginReq{
			Name: this.name,
		}
		err := this.AsynCall(login, func(i interface{}, e error) {
			if e != nil {
				session.Close(e.Error())
				return
			}
			resp := i.(*protocol.LoginResp)
			if resp.GetMsg() != "" {
				session.Close(resp.GetMsg())
				panic(resp.GetMsg())
			} else {
				this.state = state_ok
			}
		})
		if err != nil {
			session.Close(err.Error())
		}

	})

}

func (this *Launcher) dispatchMsg(session dnet.Session, msg *net.Message) {
	cmd := msg.GetCmd()
	h, ok := this.handler[cmd]
	if ok {
		h(session, msg)
	}
}

func Launch(name_, masterAddr_ string) {

	loadExecInfo()

	launcher := &Launcher{
		state:      state_none,
		session:    nil,
		name:       name_,
		masterAddr: masterAddr_,
		eventQue:   event.NewEventQueue(10240),
		handler:    map[uint16]func(session dnet.Session, msg *net.Message){},
		rpcServer:  drpc.NewServer(),
		rpcClient:  drpc.NewClient(),
	}

	launcher.handler[protocol.CmdFile] = onSyncFile
	launcher.rpcServer.Register(pb.GetNameById("pmp_req", protocol.CmdStart), onCmdStart)
	launcher.rpcServer.Register(pb.GetNameById("pmp_req", protocol.CmdSignal), onCmdSignal)

	launcher.eventQue.Run()
	launcher.dial()

}
