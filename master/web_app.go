package master

import (
	"encoding/json"
	"github.com/yddeng/dnet/dhttp"
	"github.com/yddeng/pmp/core"
	"github.com/yddeng/pmp/protocol"
	"github.com/yddeng/pmp/util"
	"net/http"
	"strings"
)

func WebAppStart() {
	conf := getConfig()

	hServer := dhttp.NewHttpServer(conf.WebApp)

	//跨域
	header := http.Header{}
	header.Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	header.Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	header.Set("content-type", "application/json")             //返回数据格式是json
	hServer.SetResponseWriterHeader(&header)

	hServer.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./app"))))
	hServer.Handle("/shared/", http.StripPrefix("/shared/", http.FileServer(http.Dir(core.FileSyncPath))))

	hServer.HandleFuncUrlParam("/script/get", scriptGet)
	hServer.HandleFuncJson("/script/create", &script{}, scriptCreate)
	hServer.HandleFuncJson("/script/update", &script{}, scriptUpdate)
	hServer.HandleFuncJson("/script/delete", &script{}, scriptDelete)

	hServer.HandleFuncUrlParam("/node/get", nodeGet)

	hServer.HandleFuncUrlParam("/item/get", itemGet)
	hServer.HandleFuncJson("/item/create", &item{}, itemCreate)
	hServer.HandleFuncJson("/item/delete", &item{}, itemDelete)

	hServer.HandleFuncJson("/itemCmd/start", &itemCmd{}, itemCmdStart)
	hServer.HandleFuncJson("/itemCmd/stop", &itemCmd{}, itemCmdStop)
	hServer.HandleFuncJson("/itemCmd/kill", &itemCmd{}, itemCmdKill)
	hServer.HandleFuncJson("/itemCmd/signal", &itemCmd{}, itemCmdSignal)

	if err := hServer.Listen(); err != nil {
		util.Logger().Errorf(err.Error())
	}
}

type resultCode struct {
	Ok      bool   `json:"ok,omitempty"`
	Message string `json:"message,omitempty"`
}

func respResult(w http.ResponseWriter, ok bool, message string) {
	ret := &resultCode{
		Ok:      ok,
		Message: message,
	}
	if err := json.NewEncoder(w).Encode(ret); err != nil {
		util.Logger().Errorf(err.Error())
	}
}

type resultData struct {
	Ok    bool        `json:"ok,omitempty"`
	Total int         `json:"total,omitempty"`
	Count int         `json:"count,omitempty"`
	Data  interface{} `json:"data,omitempty"`
}

func respData(w http.ResponseWriter, ok bool, total, count int, data interface{}) {
	ret := &resultData{
		Ok:    ok,
		Total: total,
		Count: count,
		Data:  data,
	}
	if err := json.NewEncoder(w).Encode(ret); err != nil {
		util.Logger().Errorf(err.Error())
	}
}

type script struct {
	ID   int32  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Args string `json:"args,omitempty"`
}

func scriptGet(w http.ResponseWriter, msg interface{}) {
	scripts := scriptPtr.getAll()
	total := len(scripts)
	respData(w, true, total, total, scripts)
}

func scriptCreate(w http.ResponseWriter, msg interface{}) {
	req := msg.(*script)
	req.ID = scriptPtr.genID()
	req.Args = strings.ReplaceAll(req.Args, "&", " ")
	scriptPtr.set(req.ID, req)
	respResult(w, true, "")
}

func scriptUpdate(w http.ResponseWriter, msg interface{}) {
	req := msg.(*script)
	_, ok := scriptPtr.get(req.ID)
	if !ok {
		respResult(w, false, "script not exist")
		return
	}
	req.Args = strings.ReplaceAll(req.Args, "&", " ")
	scriptPtr.set(req.ID, req)
	respResult(w, true, "")
}

func scriptDelete(w http.ResponseWriter, msg interface{}) {
	req := msg.(*script)
	if _, ok := scriptPtr.get(req.ID); !ok {
		respResult(w, false, "script not exist")
		return
	}
	scriptPtr.delete(req.ID)
	respResult(w, true, "")
}

func nodeGet(w http.ResponseWriter, msg interface{}) {
	nodes := slavePtr.getAll()
	total := len(nodes)
	respData(w, true, total, total, nodes)
}

type item struct {
	ID     int32  `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	Script int32  `json:"script,omitempty"`
	Node   string `json:"node,omitempty"`
}

func itemGet(w http.ResponseWriter, msg interface{}) {
	items := itemPtr.getAll()
	total := len(items)
	respData(w, true, total, total, items)
}

func itemCreate(w http.ResponseWriter, msg interface{}) {
	req := msg.(*item)
	if _, ok := scriptPtr.get(req.Script); !ok {
		respResult(w, false, "script not exist")
		return
	}
	if _, ok := slavePtr.get(req.Node); !ok {
		respResult(w, false, "slave not exist")
		return
	}

	req.ID = itemPtr.genID()
	itemPtr.set(req.ID, req)
	respResult(w, true, "")
}

func itemDelete(w http.ResponseWriter, msg interface{}) {
	req := msg.(*item)
	if _, ok := itemPtr.get(req.ID); !ok {
		respResult(w, false, "item not exist")
		return
	}
	itemPtr.delete(req.ID)
	respResult(w, true, "")
}

type itemCmd struct {
	ID     int32  `json:"id,omitempty"`
	Signal string `json:"signal,omitempty"`
}

func itemCmdStart(w http.ResponseWriter, msg interface{}) {
	req := msg.(*itemCmd)
	item, ok := itemPtr.get(req.ID)
	if !ok {
		respResult(w, false, "item not exist")
		return
	}
	scrip, ok := scriptPtr.get(item.Script)
	if !ok {
		respResult(w, false, "script not exist")
		return
	}
	s, ok := slavePtr.get(item.Node)
	if !ok {
		respResult(w, false, "slave not exist")
		return
	}
	start := &protocol.StartReq{
		Args:   scrip.Args,
		ItemID: item.ID,
	}
	resp, err := s.SyncCall(start)
	if err != nil {
		respResult(w, false, err.Error())
		return
	}
	ret := resp.(*protocol.StartResp)
	if ret.GetMsg() != "" {
		respResult(w, false, ret.GetMsg())
	} else {
		respResult(w, true, "")
	}
}

func itemCmdStop(w http.ResponseWriter, msg interface{}) {
	req := msg.(*itemCmd)
	item, ok := itemPtr.get(req.ID)
	if !ok {
		respResult(w, false, "item not exist")
		return
	}
	s, ok := slavePtr.get(item.Node)
	if !ok {
		respResult(w, false, "slave not exist")
		return
	}
	start := &protocol.SignalReq{
		ItemID: item.ID,
		Signal: protocol.Signal_term,
	}
	resp, err := s.SyncCall(start)
	if err != nil {
		respResult(w, false, err.Error())
		return
	}
	ret := resp.(*protocol.SignalResp)
	if ret.GetMsg() != "" {
		respResult(w, false, ret.GetMsg())
	} else {
		respResult(w, true, "")
	}
}

func itemCmdKill(w http.ResponseWriter, msg interface{}) {
	req := msg.(*itemCmd)
	item, ok := itemPtr.get(req.ID)
	if !ok {
		respResult(w, false, "item not exist")
		return
	}
	util.Logger().Infoln(item)
	respResult(w, true, "")
}

func itemCmdSignal(w http.ResponseWriter, msg interface{}) {
	req := msg.(*itemCmd)
	item, ok := itemPtr.get(req.ID)
	if !ok {
		respResult(w, false, "item not exist")
		return
	}
	sig, ok := signals[req.Signal]
	if !ok {
		respResult(w, false, "signal not exist")
		return
	}
	util.Logger().Infoln(item, sig)
	respResult(w, true, "")
}
