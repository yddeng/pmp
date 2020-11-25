package main

import (
	"github.com/yddeng/dutil/io"
	"github.com/yddeng/pmp/util"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
)

type config struct {
	Addr string
	Msg  string
}

var (
	filePath string
	conf     atomic.Value
)

func loadConfig() {
	conf_ := &config{}
	err := io.DecodeJsonFromFile(conf_, filePath)
	if err != nil {
		util.Logger().Errorln(err)
		panic(err)
	}
	conf.Store(conf_)
}

func getConfig() *config {
	return conf.Load().(*config)
}

func main() {
	if len(os.Args) < 2 {
		panic("args failed")
	}
	util.InitLogger("log", "echo")

	filePath = os.Args[1]
	util.Logger().Infoln("filePath", filePath)
	loadConfig()

	conf_ := getConfig()
	util.Logger().Infoln("echo", conf_.Addr, conf_.Msg)

	http.HandleFunc("/echo", func(writer http.ResponseWriter, request *http.Request) {
		conf_ := getConfig()
		writer.Write([]byte(conf_.Msg))
	})

	go func() {
		if err := http.ListenAndServe(conf_.Addr, nil); err != nil {
			util.Logger().Errorln(err)
		}
	}()

	for {
		select {
		case <-ListenStop():
			util.Logger().Infoln("signal stop")
			return
		case <-ListenUser1():
			util.Logger().Infoln("signal user1")
			loadConfig()
		}
	}
}

func ListenStop() <-chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)
	return sigChan
}

func ListenUser1() <-chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGUSR1)
	return sigChan
}

func ListenUser2() <-chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGUSR2)
	return sigChan
}
