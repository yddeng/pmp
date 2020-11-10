package master

import (
	"github.com/yddeng/dnet"
	"github.com/yddeng/dnet/drpc"
	"github.com/yddeng/dnet/dtcp"
	"github.com/yddeng/dutil/event"
	"github.com/yddeng/pmp/net"
	"github.com/yddeng/pmp/net/pb"
	"github.com/yddeng/pmp/protocol"
	"github.com/yddeng/pmp/util"
)

var (
	eventQueue *event.EventQueue
	hanlder    map[uint16]func(slave *Slave, msg *net.Message)
	rpcServer  *drpc.Server
	rpcClient  *drpc.Client
)

func startListener() error {
	conf := getConfig()

	addr := conf.Service
	l, err := dtcp.NewTCPListener("tcp", addr)
	if err != nil {
		return err
	}

	return l.Listen(func(session dnet.Session) {
		eventQueue.Push(func() {
			// 超时时间
			//session.SetTimeout(time.Second*10, 0)
			session.SetCodec(net.NewCodec("pmp_msg", "pmp_req", "pmp_resp"))
			session.SetCloseCallBack(onClientClose)

			_ = session.Start(func(data interface{}, err error) {
				if err != nil {
					util.Logger().Errorln(err.Error())
					session.Close(err.Error())
				} else {
					eventQueue.Push(func() {
						switch data.(type) {
						case *drpc.Request:
							rpcServer.OnRPCRequest(&Slave{session: session}, data.(*drpc.Request))
						case *drpc.Response:
							rpcClient.OnRPCResponse(data.(*drpc.Response))
						case *net.Message:
							dispatchMsg(session, data.(*net.Message))
						}
					})
				}
			})

		})

	})
}

func dispatchMsg(session dnet.Session, msg *net.Message) {
	ctx := session.Context()
	if ctx == nil {
		return
	}
	slave := ctx.(*Slave)

	cmd := msg.GetCmd()
	h, ok := hanlder[cmd]
	if !ok {
		return
	}

	h(slave, msg)

}

func initHandler() {
	hanlder = map[uint16]func(slave *Slave, msg *net.Message){}
	rpcServer.Register(pb.GetNameById("pmp_req", protocol.CmdLogin), onLogin)
}

func Launch() {
	rpcServer = drpc.NewServer()
	rpcClient = drpc.NewClient()

	slavePtr = &slave{slaves: map[string]*Slave{}}
	eventQueue = event.NewEventQueue(10240)
	eventQueue.Run()

	initHandler()
	loadScriptFile()
	loadItemFile()

	if err := startListener(); err != nil {
		panic(err)
	}
}
