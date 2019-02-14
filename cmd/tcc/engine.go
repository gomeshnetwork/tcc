package main

import (
	config "github.com/dynamicgo/go-config"
	_ "github.com/dynamicgo/slf4go-aliyun"
	_ "github.com/gomeshnetwork/agent/basic"
	"github.com/gomeshnetwork/gomesh"
	"github.com/gomeshnetwork/gomesh/app"
	"github.com/gomeshnetwork/tcc/engine/services/scheduler"
)

func main() {

	gomesh.LocalService("tcc.Scheduler", func(config config.Config) (gomesh.Service, error) {
		return scheduler.New(config)
	})

	app.Run("tcc")
}
