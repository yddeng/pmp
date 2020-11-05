package slave

import (
	"github.com/yddeng/dnet"
	"github.com/yddeng/pmp/net"
	"github.com/yddeng/pmp/protocol"
	"github.com/yddeng/pmp/util"
	"io"
	"os"
	"path"
)

var writeFiles = map[string]*writeFile{}

type writeFile struct {
	f    *os.File
	name string
}

func onSyncFile(session dnet.Session, msg *net.Message) {
	req := msg.GetData().(*protocol.File)
	if !msg.Next {
		// 表示所有数据已经发送完毕
		util.Logger().Infof("sync all file end\n")
		return
	}
	filename := req.GetFileName()

	wf, ok := writeFiles[filename]
	if !ok {
		wf = &writeFile{name: filename}
		_ = os.MkdirAll(path.Dir(filename), os.ModePerm)
		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
		if err != nil {
			panic(err)
		}
		wf.f = f
		writeFiles[filename] = wf
	}

	data := req.GetB()
	length := int(req.GetLength())
	total := 0
	for total < length {
		n, err := wf.f.Write(data[total:])
		if err != nil {
			if err != io.ErrShortWrite {
				util.Logger().Errorf(err.Error())
				break
			}
		} else {
			total += n
		}
	}

	if !req.GetNext() {
		wf.f.Close()
		delete(writeFiles, filename)
		util.Logger().Infof("sync file %s write end\n", filename)
	}

}
