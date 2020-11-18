package master

import (
	"encoding/json"
	"fmt"
	"github.com/yddeng/dnet/dhttp"
	"github.com/yddeng/pmp/core"
	"github.com/yddeng/pmp/protocol"
	"github.com/yddeng/pmp/util"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"
)

func WebAppStart() {
	conf := getConfig()

	hServer := dhttp.NewHttpServer(conf.WebApp)

	webAddr := fmt.Sprintf(`var httpAddr = "http://%s";`, conf.WebApp)
	err := ioutil.WriteFile("./app/js/addr.js", []byte(webAddr), os.ModePerm)
	if err != nil {
		panic(err)
	}

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

	hServer.HandleFuncUrlParam("/file/get", fileGet)
	hServer.HandleFuncUrlParam("/file/delete", fileDelete)
	hServer.HandleFuncUrlParam("/file/download", fileDownload)
	hServer.HandleFunc("/file/update", fileUpdate)

	if err := hServer.Listen(); err != nil {
		util.Logger().Errorf(err.Error())
	}
}

type resultCode struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
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

type item struct {
	ID      int32  `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Script  int32  `json:"script,omitempty"`
	Slave   string `json:"slave,omitempty"`
	Date    string `json:"date,omitempty"`
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
	req.Date = time.Now().Format(core.TimeFormat)
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

/***************************** 项目操作 end ******************************************/

/***************************** 通知 start ******************************************/
type notify struct {
	Type string `json:"type"`
	Url  string `json:"url"`
}

/***************************** 通知 end ******************************************/

/***************************** 文件管理 start ******************************************/

type fileNode struct {
	Filename string `json:"filename"`
	IsDir    bool   `json:"is_dir"`
	Size     int64  `json:"size"`
	Date     string `json:"date"`
}

/*
 * 获取目录下文件， 正在上传的文件不显示。
 * path -> 获取文件路径
 */
func fileGet(w http.ResponseWriter, msg interface{}) {
	req := msg.(url.Values)
	filePath := req.Get("path")
	util.Logger().Debugln("fileGet", filePath)

	info, ok := filePtr.filePath(filePath, false)
	if !ok {
		respData(w, false, 0, 0, nil)
		return
	}

	data := map[string]fileNode{}
	for _, info := range info.FileInfos {
		// 正在上传中的文件不同步
		if info.UploadInfo == nil {
			data[info.Name] = fileNode{
				Filename: info.Name,
				IsDir:    info.IsDir,
				Size:     info.Size,
				Date:     info.Date,
			}
		}
	}
	respData(w, true, len(data), len(data), data)
}

/*
 * 删除文件，文件夹。
 * path -> 文件路径
 * filename -> 文件名，文件夹名。
 */
func fileDelete(w http.ResponseWriter, msg interface{}) {
	req := msg.(url.Values)
	filePath := req.Get("path")
	filename := req.Get("filename")
	util.Logger().Debugln("fileDelete", filePath, filename)

	if filename == "" {
		respResult(w, false, "filename is nil")
		return
	}

	info, ok := filePtr.filePath(filePath, false)
	if !ok {
		respData(w, false, 0, 0, nil)
		return
	}

	filePtr.mtx.RLock()
	file, ok := info.FileInfos[filename]
	fileAbs := path.Join(file.Path, file.Name)
	filePtr.mtx.RUnlock()
	if !ok {
		respResult(w, false, "filename is not exist")
		return
	}

	// 删除本地文件
	util.Logger().Debugln("fileDelete", fileAbs)
	if err := os.RemoveAll(fileAbs); err != nil {
		util.Logger().Errorln(err)
	}

	filePtr.mtx.Lock()
	delete(info.FileInfos, filename)
	writeFileFile()
	filePtr.mtx.Unlock()

	respResult(w, true, "")
}

type fileUpdateResp struct {
	Code   int      `json:"code"` // 0->ok, 1-> 操作失败, 2->需要上传,3->不需要上传 ,
	Upload []string `json:"upload"`
}

func respFileUpdate(w http.ResponseWriter, code int, up []string) {
	ret := &fileUpdateResp{
		Code:   code,
		Upload: up,
	}
	if err := json.NewEncoder(w).Encode(ret); err != nil {
		util.Logger().Errorf(err.Error())
	}
}

/*
 * 文件上传，创建路径。
 * path -> 文件路径
 * filename -> 文件名。当filename为空时，仅创建文件夹。
 * file -> 文件分片。
 * total -> 文件总分片数。
 * current -> 当前文件分片。当 current == 0，验证文件存在，断点续传。
 * md5 -> 文件md5值。比对文件变化。
 */
func fileUpdate(w http.ResponseWriter, r *http.Request) {
	filePath := r.FormValue("path")
	filename := r.FormValue("filename")

	util.Logger().Infoln("fileUpdate", filePath, filename, r.Form)

	info, ok := filePtr.filePath(filePath, true)
	if !ok {
		respFileUpdate(w, 1, nil)
		return
	}

	if filename == "" {
		respFileUpdate(w, 0, nil)
		return
	}

	md5 := r.FormValue("md5")
	current := r.FormValue("current")
	total := r.FormValue("total")
	totalInt, err := strconv.Atoi(total)
	if err != nil {
		respFileUpdate(w, 1, nil)
		return
	}

	filePtr.mtx.RLock()
	file, ok := info.FileInfos[filename]
	// 判断是否有分片已经上传
	if current == "0" {
		if !ok {
			respFileUpdate(w, 2, nil)
		} else {
			// 对比md5,不同
			if file.MD5 != md5 {
				// 清理上传的分片
				if file.UploadInfo != nil {
					for id := range file.UploadInfo.UpLoad {
						filePtr.mtx.RUnlock()
						tmpFilename := file.makeSliceFilename(id)
						err := os.Remove(tmpFilename)
						if err != nil {
							util.Logger().Errorf(err.Error())
						}
						filePtr.mtx.RLock()
					}
				}
				fileAbs := path.Join(file.Path, file.Name)
				filePtr.mtx.RUnlock()

				// 清理文件
				err := os.Remove(fileAbs)
				if err != nil {
					util.Logger().Errorf(err.Error())
				}

				filePtr.mtx.Lock()
				delete(info.FileInfos, filename)
				writeFileFile()
				filePtr.mtx.Unlock()

				respFileUpdate(w, 2, nil)
			} else {
				// 文件已经上传成功
				if file.UploadInfo == nil {
					filePtr.mtx.RUnlock()
					respFileUpdate(w, 3, nil)
				} else {
					// 切片数量改变
					if file.UploadInfo.Total != totalInt {
						filePtr.mtx.RUnlock()
						filePtr.mtx.Lock()
						file.UploadInfo = nil
						writeFileFile()
						filePtr.mtx.Unlock()
						respFileUpdate(w, 2, nil)
					} else {
						// 已有分片存在
						exist := []string{}
						for id := range file.UploadInfo.UpLoad {
							exist = append(exist, id)
						}
						respFileUpdate(w, 2, exist)
						filePtr.mtx.RUnlock()
					}
				}
			}
			filePtr.mtx.RLock()
		}
		filePtr.mtx.RUnlock()
		return
	}

	if !ok {
		file = &fileInfo{
			Path: path.Join(info.Path, info.Name),
			Name: filename,
			MD5:  md5,
			UploadInfo: &uploadInfo{
				Total:  totalInt,
				UpLoad: map[string]struct{}{},
			},
		}
		filePtr.mtx.RUnlock()
		filePtr.mtx.Lock()
		info.FileInfos[file.Name] = file
		writeFileFile()
		filePtr.mtx.Unlock()
		filePtr.mtx.RLock()
	}

	defer file.tryMerge()

	_, ok = file.UploadInfo.UpLoad[current]
	if ok {
		// 当前分片已经上传
		respFileUpdate(w, 0, nil)
		filePtr.mtx.RUnlock()
		return
	}
	filePtr.mtx.RUnlock()

	gFile, _, err := r.FormFile("file")
	if err != nil {
		util.Logger().Errorln(err)
		respFileUpdate(w, 1, nil)
		return
	}
	defer gFile.Close()

	tmpFilename := file.makeSliceFilename(current)
	if err = util.WriteFile(tmpFilename, gFile); err != nil {
		util.Logger().Debugln(err.Error())
		respFileUpdate(w, 1, nil)
		return
	}

	filePtr.mtx.Lock()
	file.UploadInfo.UpLoad[current] = struct{}{}
	writeFileFile()
	filePtr.mtx.Unlock()

	respFileUpdate(w, 0, nil)

}

/*
 * 文件下载
 * path -> 文件路径
 * filename -> 文件名。
 */
func fileDownload(w http.ResponseWriter, msg interface{}) {
	req := msg.(url.Values)
	filePath := req.Get("path")
	filename := req.Get("filename")
	util.Logger().Debugln("fileDownload", filePath, filename)

	//打开文件
	fileAbs := path.Join(filePath, filename)
	file, err := os.Open(fileAbs)
	if err != nil {
		util.Logger().Errorln(err)
		respResult(w, false, fileAbs+" not exist")
		return
	}
	//结束后关闭文件
	defer file.Close()

	//设置响应的header头
	w.Header().Add("Content-type", "application/octet-stream")
	w.Header().Add("content-disposition", "attachment; filename=\""+filename+"\"")
	//将文件写至responseBody
	_, err = io.Copy(w, file)
	if err != nil {
		util.Logger().Errorln(err)
		respResult(w, false, err.Error())
		return
	}
}

/***************************** 文件管理 end ******************************************/
