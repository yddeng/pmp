package master

import (
	"github.com/yddeng/pmp/core"
	"github.com/yddeng/pmp/net"
	"github.com/yddeng/pmp/protocol"
	"io/ioutil"
	"os"
	"path/filepath"
)

func getAndSyncAll(slave *Slave) error {
	length := net.BuffSize - net.HeadSize - 200
	err := filepath.Walk(core.FileSyncPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			idx := 0
			total := len(data)
			for total > length {
				file := &protocol.File{
					FileName: path,
					B:        data[idx : idx+length],
					Next:     true,
					Length:   int32(length),
				}
				slave.SendMessage(file, true)
				idx += length
				total -= length
			}

			file := &protocol.File{
				FileName: path,
				B:        data[idx:],
				Length:   int32(total),
			}
			slave.SendMessage(file, true)

		}
		return nil
	})
	if err != nil {
		return err
	}

	slave.SendMessage(&protocol.File{})
	eventQueue.Push(func() {
		slave.ok = true
	})
	return nil
}
