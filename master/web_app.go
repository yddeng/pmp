package master

import (
	"encoding/json"
	"github.com/yddeng/dnet/dhttp"
	"github.com/yddeng/pmp/core"
	"github.com/yddeng/pmp/protocol"
	"github.com/yddeng/pmp/util"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
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
	hServer.Handle("/shared/", http.StripPrefix("/shared/", http.FileServer(http.Dir(core.SharedPath))))

	hServer.HandleFuncUrlParam("/script/get", scriptGet)
	hServer.HandleFuncJson("/script/create", &script{}, scriptCreate)
	hServer.HandleFuncJson("/script/update", &script{}, scriptUpdate)
	hServer.HandleFuncJson("/script/delete", &script{}, scriptDelete)

	hServer.HandleFuncUrlParam("/node/get", nodeGet)

	hServer.HandleFuncUrlParam("/item/get", itemGet)
	hServer.HandleFuncJson("/item/create", &item{}, itemCreate)
	hServer.HandleFuncJson("/item/delete", &item{}, itemDelete)

	hServer.HandleFuncJson("/itemCmd/start", &itemCmd{}, itemCmdStart)
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
		Ok: ok,
	}
	if ok {
		ret.Total = total
		ret.Count = count
		ret.Data = data
	}
	if err := json.NewEncoder(w).Encode(ret); err != nil {
		util.Logger().Errorf(err.Error())
	}
}

/***************************** 脚本 start ******************************************/

type script struct {
	ID   int32  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Date string `json:"date,omitempty"`
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
	req.Date = time.Now().Format(core.TimeFormat)
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
	req.Date = time.Now().Format(core.TimeFormat)
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

/***************************** 脚本 end ******************************************/

/***************************** 节点信息 start ******************************************/

type node struct {
	ID   int32             `json:"id,omitempty"`
	Name string            `json:"name,omitempty"`
	Sys  *protocol.SysInfo `json:"sys,omitempty"`
}

func nodeGet(w http.ResponseWriter, msg interface{}) {
	req := msg.(url.Values)
	n := req.Get("n")
	switch n {
	case "list":
		list := []node{}
		nodes := slavePtr.getAll()
		for _, v := range nodes {
			list = append(list, node{ID: v.id, Name: v.name})
		}
		total := len(list)
		respData(w, true, total, total, list)
	default:
		num, err := strconv.Atoi(n)
		if err != nil {
			respData(w, false, 0, 0, nil)
			return
		}

		slave, ok := slavePtr.get(int32(num))
		if !ok {
			respData(w, false, 0, 0, nil)
			return
		}

		port := slave.GetReport()
		respData(w, true, 1, 1, node{
			ID:   slave.id,
			Name: slave.name,
			Sys:  port.GetSys(),
		})
	}
}

/***************************** 节点信息 end ******************************************/

type item struct {
	ID      int32  `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Script  int32  `json:"script,omitempty"`
	Slave   int32  `json:"slave,omitempty"`
	IsGuard bool   `json:"is_guard,omitempty"`
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
	if _, ok := slavePtr.get(req.Slave); !ok {
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
	s, ok := slavePtr.get(item.Slave)
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

func itemCmdSignal(w http.ResponseWriter, msg interface{}) {
	req := msg.(*itemCmd)
	item, ok := itemPtr.get(req.ID)
	if !ok {
		respResult(w, false, "item not exist")
		return
	}
	s, ok := slavePtr.get(item.Slave)
	if !ok {
		respResult(w, false, "slave not exist")
		return
	}
	signal := &protocol.SignalReq{
		ItemID: item.ID,
	}
	switch req.Signal {
	case "term":
		signal.Signal = protocol.Signal_term
	case "kill":
		signal.Signal = protocol.Signal_kill
	case "user1":
		signal.Signal = protocol.Signal_user1
	case "user2":
		signal.Signal = protocol.Signal_user2
	default:
		respResult(w, false, "signal invalid")
		return
	}
	resp, err := s.SyncCall(signal)
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

type notify struct {
	Type string `json:"type"`
	Url  string `json:"url"`
}
