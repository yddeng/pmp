package master

import (
	"github.com/yddeng/pmp/protocol"
	"github.com/yddeng/pmp/util"
)

func Start() {
	eventQueue.Push(func() {
		slave := getSlave()
		if slave != nil {
			slave.AsynCall(&protocol.StartReq{Name: "SYNC/echo/echo", Command: "SYNC/conf/config.json"}, func(i interface{}, e error) {
				util.Logger().Infoln(i, e)
			})
		}
	})
}
