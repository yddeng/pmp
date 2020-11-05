package main

import (
	"flag"
	"github.com/yddeng/pmp/slave"
	"github.com/yddeng/pmp/util"
)

var (
	name       = flag.String("name", "pmp_slave", "pmp_slave name")
	masterAddr = flag.String("master", "127.0.0.1:23456", "master addr")
)

func main() {
	util.InitLogger("log", "slave")
	slave.Launch(*name, *masterAddr)
	select {}
}
