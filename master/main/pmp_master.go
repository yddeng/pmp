package main

import (
	"github.com/yddeng/pmp/core"
	"github.com/yddeng/pmp/master"
	"github.com/yddeng/pmp/util"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		panic("args: config")
	}

	_ = os.MkdirAll(core.SharedPath, os.ModePerm)
	_ = os.MkdirAll(core.DataPath, os.ModePerm)

	util.InitLogger("log", "master")

	master.LoadConfig(os.Args[1])
	master.Launch()
	master.WebAppStart()

	select {}
}
