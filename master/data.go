package master

import (
	"github.com/yddeng/dutil/io"
	"github.com/yddeng/pmp/core"
	"github.com/yddeng/pmp/util"
	"path"
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
