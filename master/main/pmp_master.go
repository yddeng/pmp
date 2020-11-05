package main

import (
	"github.com/yddeng/pmp/core"
	"github.com/yddeng/pmp/master"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		panic("args: config")
	}

	os.MkdirAll(core.FileSyncPath, os.ModePerm)

	master.LoadConfig(os.Args[1])
	master.Launch()

	select {}
}
