package master

import (
	"fmt"
	"github.com/yddeng/dutil/io"
	"github.com/yddeng/gsf/util/time"
	"github.com/yddeng/pmp/core"
	"github.com/yddeng/pmp/util"
	io2 "io"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
)

const (
	scriptDataName = "script.json"
	itemDataName   = "item.json"
)

var (
	scriptPtr      *scriptFile
	scriptFilename string
	itemPtr        *itemFile
	itemFilename   string
	signals        map[string]int32
)

type scriptFile struct {
	mtx         sync.RWMutex      `json:"-"`
	Scripts     map[int32]*script `json:"scripts"`
	ScriptGenID int32             `json:"script_gen_id"`
}

func loadScriptFile() {
	scriptFilename = path.Join(core.DataPath, scriptDataName)
	scriptPtr = &scriptFile{
		Scripts:     map[int32]*script{},
		ScriptGenID: 0,
	}
	if err := io.DecodeJsonFromFile(scriptPtr, scriptFilename); err != nil {
		util.Logger().Errorf(err.Error())
	}
}

func writeScriptFile() {
	if err := io.EncodeJsonToFile(scriptPtr, scriptFilename); err != nil {
		util.Logger().Errorf(err.Error())
	}
}

func (this *scriptFile) genID() int32 {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	this.ScriptGenID++
	writeScriptFile()
	return this.ScriptGenID
}

func (this *scriptFile) getAll() map[int32]*script {
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	return this.Scripts
}

func (this *scriptFile) set(key int32, val *script) {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	this.Scripts[key] = val
	writeScriptFile()
}

func (this *scriptFile) get(key int32) (s *script, ok bool) {
	this.mtx.RLock()
	s, ok = this.Scripts[key]
	this.mtx.RUnlock()
	return
}

func (this *scriptFile) delete(key int32) {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	delete(this.Scripts, key)
	writeScriptFile()
}

type itemFile struct {
	mtx       sync.RWMutex    `json:"-"`
	Items     map[int32]*item `json:"items"`
	ItemGenID int32           `json:"item_gen_id"`
}

func loadItemFile() {
	itemFilename = path.Join(core.DataPath, itemDataName)
	itemPtr = &itemFile{
		Items:     map[int32]*item{},
		ItemGenID: 0,
	}
	if err := io.DecodeJsonFromFile(itemPtr, itemFilename); err != nil {
		util.Logger().Errorf(err.Error())
	}
}

func writeItemFile() {
	if err := io.EncodeJsonToFile(itemPtr, itemFilename); err != nil {
		util.Logger().Errorf(err.Error())
	}
}

func (this *itemFile) genID() int32 {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	this.ItemGenID++
	writeItemFile()
	return this.ItemGenID
}

func (this *itemFile) getAll() map[int32]*item {
	this.mtx.RLock()
	defer this.mtx.RUnlock()
	return this.Items
}

func (this *itemFile) set(key int32, val *item) {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	this.Items[key] = val
	writeItemFile()
}

func (this *itemFile) get(key int32) (s *item, ok bool) {
	this.mtx.RLock()
	s, ok = this.Items[key]
	this.mtx.RUnlock()
	return
}

func (this *itemFile) delete(key int32) {
	this.mtx.Lock()
	defer this.mtx.Unlock()
	delete(this.Items, key)
	writeItemFile()
}

// 文件管理

var (
	filePtr      *fileFile
	fileInfoName = "fileInfo.json"
	fileFilename string
)

type fileFile struct {
	mtx      sync.RWMutex `json:"-"`
	FileInfo *fileInfo    `json:"file_info"`
}

func (this *fileFile) splitDir(dir string) []string {
	paths := strings.Split(dir, "/")
	l := []string{}
	for _, v := range paths {
		if v != "" {
			l = append(l, v)
		}
	}
	return l
}

func (this *fileFile) filePath(filePath string, mkdir bool) (*fileInfo, bool) {
	paths := this.splitDir(filePath)
	//if len(paths) < 1 {
	//	return nil, false
	//}

	filePtr.mtx.Lock()
	defer filePtr.mtx.Unlock()

	info := filePtr.FileInfo
	for i := 1; i < len(paths); i++ {
		dname := paths[i]
		cInfo, ok := info.FileInfos[dname]
		if ok {
			if !cInfo.IsDir {
				return nil, false
			}
		} else {
			cInfo = &fileInfo{
				Path:      path.Join(info.Path, info.Name),
				Name:      dname,
				IsDir:     true,
				FileInfos: map[string]*fileInfo{},
			}
			_ = os.MkdirAll(path.Join(cInfo.Path, cInfo.Name), os.ModePerm)
			info.FileInfos[cInfo.Name] = cInfo
			writeFileFile()
		}
		info = cInfo
	}
	return info, true
}

type fileInfo struct {
	Path       string               `json:"path"`
	Name       string               `json:"name,omitempty"`
	IsDir      bool                 `json:"is_dir,omitempty"`
	Size       int64                `json:"size,omitempty"`       // 文件夹为0
	MD5        string               `json:"md5,omitempty"`        // 文价夹为空
	Date       string               `json:"date,omitempty"`       // 文价夹为空
	FileInfos  map[string]*fileInfo `json:"file_info"`            // dir -> fileInfo 。 只有是文件夹才有值
	UploadInfo *uploadInfo          `json:"uploadInfo,omitempty"` // 为空时，表示文件传输完成已合并文件
}

func (this *fileInfo) makeSliceFilename(crt string) string {
	return fmt.Sprintf("%s.part%s", path.Join(this.Path, this.Name), crt)
}

type uploadInfo struct {
	Total  int                 `json:"total"`
	UpLoad map[string]struct{} `json:"up_load"` // 已经上传的切片
}

func (this *fileInfo) tryMerge() {
	filePtr.mtx.RLock()
	// 合并分片
	if this.UploadInfo == nil {
		filePtr.mtx.RUnlock()
		return
	}

	if this.UploadInfo.Total == len(this.UploadInfo.UpLoad) {
		filename := path.Join(this.Path, this.Name)
		filePtr.mtx.RUnlock()
		f, err := os.Create(filename)
		if err != nil {
			util.Logger().Errorf(err.Error())
			return
		}
		size := int64(0)
		filePtr.mtx.RLock()
		for i := 1; i <= this.UploadInfo.Total; i++ {
			tmpFilename := this.makeSliceFilename(strconv.Itoa(i))
			filePtr.mtx.RUnlock()
			tf, err := os.Open(tmpFilename)
			if err != nil {
				util.Logger().Errorf(err.Error())
				return
			}

			written, err := io2.Copy(f, tf)
			_ = tf.Close()
			if err != nil {
				util.Logger().Errorf(err.Error())
				return
			}

			_ = os.Remove(tmpFilename)
			size += written
			util.Logger().Infof("input %s from %s written %d", this.Name, tmpFilename, written)
			filePtr.mtx.RLock()
		}
		_ = f.Close()

		filePtr.mtx.RUnlock()
		filePtr.mtx.Lock()
		this.UploadInfo = nil
		this.Date = time.Now().Format(core.TimeFormat)
		this.Size = size
		writeFileFile()
		filePtr.mtx.Unlock()
		return
	}

	filePtr.mtx.RUnlock()
}

func loadFileFile() {
	fileFilename = path.Join(core.DataPath, fileInfoName)
	filePtr = &fileFile{
		mtx: sync.RWMutex{},
		FileInfo: &fileInfo{
			Path:      "",
			Name:      core.SharedPath,
			IsDir:     true,
			FileInfos: map[string]*fileInfo{},
		},
	}

	if err := io.DecodeJsonFromFile(filePtr, fileFilename); err != nil {
		util.Logger().Errorf(err.Error())
	}
}

func writeFileFile() {
	if err := io.EncodeJsonToFile(filePtr, fileFilename); err != nil {
		util.Logger().Errorf(err.Error())
	}
}
