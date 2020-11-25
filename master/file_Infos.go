package master

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/yddeng/gsf/util/time"
	"github.com/yddeng/pmp/core"
	"github.com/yddeng/pmp/util"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

func makeFilePart(name, part string) string {
	return fmt.Sprintf("%s.part%s", name, part)
}

func splitPath(dir string) []string {
	paths := strings.Split(dir, "/")
	l := []string{}
	for _, v := range paths {
		if v != "" {
			l = append(l, v)
		}
	}
	return l
}

// 文件 md5 值
func fileMD5(filename string) (string, error) {
	h := md5.New()
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(h, f)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

var (
	filePtr *fileInfos
)

type fileInfos struct {
	mtx      sync.RWMutex        `json:"-"`
	FileInfo *fileInfo           `json:"file_info"`
	MD5Files map[string]*md5File `json:"md_5_files"`
}

type md5File struct {
	File string   `json:"file"` // 原始文件
	Size int64    `json:"file_size"`
	MD5  string   `json:"md_5"`
	Ptr  []string `json:"ptr"` // 文件引用
}

type fileInfo struct {
	Path      string               `json:"path"`           // 相对路径
	Name      string               `json:"name,omitempty"` // 名字
	AbsPath   string               `json:"abs_path"`       // 绝对路径
	IsDir     bool                 `json:"is_dir,omitempty"`
	FileOk    bool                 `json:"file_ok"`    // 当前文件是否已经写入
	FileSize  int64                `json:"file_size"`  // 文件有值
	FileMD5   string               `json:"file_md_5"`  // 文件有值
	FileDate  string               `json:"file_date"`  // 文件有值
	FileInfos map[string]*fileInfo `json:"file_infos"` // 文件夹有值
	Upload    *upload
}

type upload struct {
	Size     int64             `json:"size,omitempty"` // 文件有值
	MD5      string            `json:"md5,omitempty"`  // 文件有值
	SliceCnt int               `json:"slice_cnt"`      // 文件有值，文件上传时总文件数。
	UpSlice  map[string]string `json:"up_slice"`       // 文件有值，文件上传时，已经上传的分片
}

func (this *fileInfo) clearUpload() {
	if this.Upload != nil {
		for part := range this.Upload.UpSlice {
			filename := makeFilePart(this.AbsPath, part)
			_ = os.RemoveAll(filename)
		}
	}
}

func (this *fileInfo) mergeUpload() {
	if this.Upload == nil {
		return
	}
	if this.Upload.SliceCnt != len(this.Upload.UpSlice) {
		return
	}
	f, err := os.Create(this.AbsPath)
	if err != nil {
		util.Logger().Errorln(err)
		return
	}
	defer f.Close()

	for i := 1; i <= this.Upload.SliceCnt; i++ {
		partFile := makeFilePart(this.AbsPath, strconv.Itoa(i))
		pf, err := os.Open(partFile)
		if err != nil {
			util.Logger().Errorln(err)
			return
		}
		written, err := io.Copy(f, pf)
		_ = pf.Close()
		if err != nil {
			util.Logger().Errorln(err)
			return
		}
		util.Logger().Infof("input %s from %s written %d ", this.AbsPath, partFile, written)
	}

	this.clearUpload()

	this.FileOk = true
	this.FileSize = this.Upload.Size
	this.FileMD5 = this.Upload.MD5
	this.FileDate = time.Now().Format(core.TimeFormat)
	this.Upload = nil

	filePtr.addMD5File(this.FileMD5, this)
}

func (this *fileInfos) addMD5File(md5 string, info *fileInfo) {
	files, ok := this.MD5Files[md5]
	if !ok {
		files = &md5File{
			File: info.AbsPath,
			MD5:  info.FileMD5,
			Size: info.FileSize,
			Ptr:  []string{},
		}
		this.MD5Files[md5] = files
	}
	files.Ptr = append(files.Ptr, info.AbsPath)
}

func (this *fileInfos) removeMD5File(md5, ptr string) {
	// 删除md5指向
	files, ok := this.MD5Files[md5]
	if ok {
		idx := -1
		for i := 0; i < len(files.Ptr); i++ {
			if files.Ptr[i] == ptr {
				idx = i
				break
			}
		}
		if idx != -1 {
			files.Ptr = append(files.Ptr[:idx], files.Ptr[idx+1:]...)
			if len(files.Ptr) == 0 {
				delete(this.MD5Files, md5)
			}
		}
	}
}

// 文件删除，
func (this *fileInfos) remove(parent *fileInfo, name string) error {
	info, ok := parent.FileInfos[name]
	if !ok {
		return fmt.Errorf("%s 文件不存在", name)
	}

	delMd5 := map[string]struct{}{} // 待删除的md5文件，源文件

	// 遍历文件
	if err := walk(info, func(file *fileInfo) error {
		if !config.SaveFileMultiple {
			if md5File_, ok := filePtr.MD5Files[file.FileMD5]; ok {
				if md5File_.File == file.AbsPath {
					// 此文件为源文件
					delMd5[file.FileMD5] = struct{}{}
				}
			}
		}

		// 删除md5指向
		this.removeMD5File(file.FileMD5, file.AbsPath)
		// 清理上传的分片
		file.clearUpload()

		return nil
	}); err != nil {
		return err
	}

	// 删除info
	delete(parent.FileInfos, info.Name)

	if !config.SaveFileMultiple {
		// 如果文件夹中包含源文件需要拷贝到他处
		for md5 := range delMd5 {
			md5File_, ok := filePtr.MD5Files[md5]
			if ok {
				// 还存在他处引用
				_ = os.Rename(md5File_.File, md5File_.Ptr[0])
				md5File_.File = md5File_.Ptr[0]
			}
		}
	}

	// 删除文件、文件夹
	if err := os.RemoveAll(info.AbsPath); err != nil {
		util.Logger().Errorln(err)
	}

	return nil
}

func (this *fileInfos) findPath(filePath string, mkdir bool) (*fileInfo, error) {
	paths := splitPath(filePath)

	info := filePtr.FileInfo
	for i := 1; i < len(paths); i++ {
		dname := paths[i]
		cInfo, ok := info.FileInfos[dname]
		if ok {
			if !cInfo.IsDir {
				return nil, fmt.Errorf("已存在同名文件！")
			}
		} else {
			cInfo = &fileInfo{
				Path:      path.Join(info.Path, info.Name),
				Name:      dname,
				AbsPath:   path.Join(info.AbsPath, dname),
				IsDir:     true,
				FileInfos: map[string]*fileInfo{},
			}
			if err := os.MkdirAll(path.Join(cInfo.Path, cInfo.Name), os.ModePerm); err != nil {
				return nil, err
			}
			info.FileInfos[cInfo.Name] = cInfo
		}
		info = cInfo
	}
	return info, nil
}

// 遍历info，调用文件
func walk(info *fileInfo, f func(file *fileInfo) error) (err error) {
	if info == nil {
		return
	}
	if !info.IsDir {
		return f(info)
	}
	for _, cInfo := range info.FileInfos {
		if cInfo.IsDir {
			err = walk(cInfo, f)
		} else {
			err = f(cInfo)
		}
		if err != nil {
			return
		}
	}
	return
}

func loadFilePath(filePath string) {
	_ = os.MkdirAll(core.SharedPath, os.ModePerm)
	sdir, dname := path.Split(filePath)
	filePtr = &fileInfos{
		mtx: sync.RWMutex{},
		FileInfo: &fileInfo{
			Path:      "",
			Name:      dname,
			AbsPath:   filePath,
			IsDir:     true,
			FileInfos: map[string]*fileInfo{},
		},
		MD5Files: map[string]*md5File{},
	}

	err := filepath.Walk(filePath, func(absPath string, f os.FileInfo, err error) error {
		if err != nil {
			util.Logger().Errorln(err)
			return err
		}

		relativePath := strings.TrimPrefix(absPath, sdir)
		if !f.IsDir() {
			// 是文件

			_, filename := path.Split(absPath)
			//fmt.Println(absPath, f.Name(), filename)
			if strings.Contains(filename, ".part") {
				// 是上传时的文件分片，删除
				_ = os.RemoveAll(absPath)
			} else {
				md5, e := fileMD5(absPath)
				if e != nil {
					util.Logger().Errorln(e)
					return e
				}
				dir, file := path.Split(relativePath)
				info, _ := filePtr.findPath(dir, true)
				fInfo := &fileInfo{
					Path:     path.Join(info.Path, info.Name),
					Name:     file,
					AbsPath:  path.Join(info.AbsPath, file),
					IsDir:    false,
					FileSize: f.Size(),
					FileMD5:  md5,
					FileDate: f.ModTime().Format(core.TimeFormat),
					FileOk:   true,
				}
				info.FileInfos[file] = fInfo
				filePtr.addMD5File(md5, fInfo)
			}
		} else {
			_, _ = filePtr.findPath(relativePath, true)
		}

		return nil
	})
	if err != nil {
		util.Logger().Errorln(err)
		panic(err)
	}

}
