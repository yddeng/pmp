package main

import (
	"flag"
	"github.com/yddeng/pmp/core"
	"github.com/yddeng/pmp/slave"
	"github.com/yddeng/pmp/util"
	"os"
)

var (
	name       = flag.String("name", "pmp_slave", "pmp_slave name")
	masterAddr = flag.String("master", "127.0.0.1:23456", "master addr")
)

func main() {
	util.InitLogger("log", "slave")

	_ = os.MkdirAll(core.SharedPath, os.ModePerm)
	_ = os.MkdirAll(core.DataPath, os.ModePerm)
	slave.Launch(*name, *masterAddr)
	select {}
}
