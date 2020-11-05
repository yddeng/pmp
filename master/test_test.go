package master

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestLaunch(t *testing.T) {
	filepath.Walk("syncDirectory", func(path string, info os.FileInfo, err error) error {
		fmt.Println(path, info.Name(), info.IsDir(), info.Size(), err)
		return nil
	})
}
