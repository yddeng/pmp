package master

import (
	"encoding/json"
	"fmt"
	"github.com/yddeng/dnet/dhttp"
	"github.com/yddeng/pmp/core"
	"github.com/yddeng/pmp/protocol"
	"github.com/yddeng/pmp/util"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"
)

func WebAppStart() {
	conf := getConfig()

	hServer := dhttp.NewHttpServer(conf.WebApp)

	addrStr := fmt.Sprintf(`var httpAddr = "http://%s";
var root = "%s";
var sliceSize = %d*1024*1024;`, config.WebApp, core.SharedPath, config.SliceSize)
	err := ioutil.WriteFile("./app/js/addr.js", []byte(addrStr), os.ModePerm)
	if err != nil {
		panic(err)
	}

	//跨域
	header := http.Header{}
	header.Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	header.Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	header.Set("content-type", "application/json")             //返回数据格式是json
	hServer.SetResponseWriterHeader(&header)

	hServer.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(config.WebIndex))))
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

	hServer.HandleFuncUrlParam("/file/list", fileList)
	hServer.HandleFuncUrlParam("/file/delete", fileDelete)
	hServer.HandleFuncUrlParam("/file/mkdir", fileMkdir)
	hServer.HandleFuncJson("/file/check", &fileCheckReq{}, fileCheck)
	hServer.HandleFunc("/file/upload", fileUpload)
	hServer.HandleFuncUrlParam("/file/download", fileDownload)

	if err := hServer.Listen(); err != nil {
		util.Logger().Errorf(err.Error())
	}
}

type resultCode struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

func respResult(w http.ResponseWriter, message string) {
	ret := &resultCode{
		Message: message,
	}
	if message == "" {
		ret.Ok = true
	}
	if err := json.NewEncoder(w).Encode(ret); err != nil {
		util.Logger().Errorf(err.Error())
	}
}

type resultData struct {
	Ok    bool        `json:"ok"`
	Total int         `json:"total"`
	Count int         `json:"count"`
	Data  interface{} `json:"data"`
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
	scriptPtr.set(req.ID, req)
	respResult(w, "")
}

func scriptUpdate(w http.ResponseWriter, msg interface{}) {
	req := msg.(*script)
	_, ok := scriptPtr.get(req.ID)
	if !ok {
		respResult(w, "script not exist")
		return
	}
	req.Date = time.Now().Format(core.TimeFormat)
	scriptPtr.set(req.ID, req)
	respResult(w, "")
}

func scriptDelete(w http.ResponseWriter, msg interface{}) {
	req := msg.(*script)
	if _, ok := scriptPtr.get(req.ID); !ok {
		respResult(w, "script not exist")
		return
	}
	scriptPtr.delete(req.ID)
	respResult(w, "")
}

/***************************** 脚本 end ******************************************/

/***************************** 节点信息 start ******************************************/

type node struct {
	Name string            `json:"name,omitempty"`
	Sys  *protocol.SysInfo `json:"sys,omitempty"`
}

func nodeGet(w http.ResponseWriter, msg interface{}) {
	req := msg.(url.Values)
	name := req.Get("name")
	switch name {
	case "list":
		list := []string{}
		nodes := slavePtr.getAll()
		for _, v := range nodes {
			list = append(list, v.name)
		}
		total := len(list)
		respData(w, true, total, total, list)
	default:
		slave, ok := slavePtr.get(name)
		if !ok {
			respData(w, false, 0, 0, nil)
			return
		}

		port := slave.GetReport()
		respData(w, true, 1, 1, node{
			Name: slave.name,
			Sys:  port.GetSys(),
		})
	}
}

/***************************** 节点信息 end ******************************************/

/***************************** 项目 start ******************************************/

type itemListResp struct {
	Item    *item        `json:"item"`
	RunInfo *itemRunInfo `json:"run_info"`
}

type itemRunInfo struct {
	Pid     int32   `json:"pid"`
	CpuUsed float64 `json:"cpuUsed"`
	MemUsed float64 `json:"memUsed"`
	Running bool    `json:"running"`
}

func itemGet(w http.ResponseWriter, msg interface{}) {
	runInfo := slavePtr.getRunInfo()
	items := itemPtr.getAll()
	itemList := map[int32]*itemListResp{}
	for id, item := range items {
		resp := &itemListResp{
			Item: item,
		}
		run, ok := runInfo[id]
		if ok {
			resp.RunInfo = &itemRunInfo{
				Pid:     run.GetPid(),
				CpuUsed: run.GetCpuUsed(),
				MemUsed: run.GetMemUsed(),
				Running: run.GetRunning(),
			}
		}
		itemList[id] = resp
	}
	total := len(itemList)
	respData(w, true, total, total, itemList)
}

type item struct {
	ID      int32  `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Script  int32  `json:"script,omitempty"`
	Slave   string `json:"slave,omitempty"`
	Date    string `json:"date,omitempty"`
	IsGuard bool   `json:"is_guard,omitempty"`
}

func itemCreate(w http.ResponseWriter, msg interface{}) {
	req := msg.(*item)
	if _, ok := scriptPtr.get(req.Script); !ok {
		respResult(w, "script not exist")
		return
	}
	if _, ok := slavePtr.get(req.Slave); !ok {
		respResult(w, "slave not exist")
		return
	}

	req.ID = itemPtr.genID()
	req.Date = time.Now().Format(core.TimeFormat)
	itemPtr.set(req.ID, req)
	respResult(w, "")
}

func itemDelete(w http.ResponseWriter, msg interface{}) {
	req := msg.(*item)
	if _, ok := itemPtr.get(req.ID); !ok {
		respResult(w, "item not exist")
		return
	}
	itemPtr.delete(req.ID)
	respResult(w, "")
}

/***************************** 项目 end ******************************************/

/***************************** 项目操作 start ******************************************/

type itemCmd struct {
	ID     int32  `json:"id,omitempty"`
	Signal string `json:"signal,omitempty"`
}

func itemCmdStart(w http.ResponseWriter, msg interface{}) {
	req := msg.(*itemCmd)
	item, ok := itemPtr.get(req.ID)
	if !ok {
		respResult(w, "item not exist")
		return
	}
	scrip, ok := scriptPtr.get(item.Script)
	if !ok {
		respResult(w, "script not exist")
		return
	}
	s, ok := slavePtr.get(item.Slave)
	if !ok {
		respResult(w, "slave not exist")
		return
	}
	start := &protocol.StartReq{
		Args:   scrip.Args,
		ItemID: item.ID,
	}
	resp, err := s.SyncCall(start)
	if err != nil {
		respResult(w, err.Error())
		return
	}
	ret := resp.(*protocol.StartResp)
	if ret.GetMsg() != "" {
		respResult(w, ret.GetMsg())
	} else {
		respResult(w, "")
	}
}

func itemCmdSignal(w http.ResponseWriter, msg interface{}) {
	req := msg.(*itemCmd)
	item, ok := itemPtr.get(req.ID)
	if !ok {
		respResult(w, "item not exist")
		return
	}
	s, ok := slavePtr.get(item.Slave)
	if !ok {
		respResult(w, "slave not exist")
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
		respResult(w, "signal invalid")
		return
	}
	resp, err := s.SyncCall(signal)
	if err != nil {
		respResult(w, err.Error())
		return
	}
	ret := resp.(*protocol.SignalResp)
	if ret.GetMsg() != "" {
		respResult(w, ret.GetMsg())
	} else {
		respResult(w, "")
	}
}

/***************************** 项目操作 end ******************************************/

/***************************** 通知 start ******************************************/
type notify struct {
	Type string `json:"type"`
	Url  string `json:"url"`
}

/***************************** 通知 end ******************************************/
