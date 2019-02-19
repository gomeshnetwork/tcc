package main

import (
	"github.com/bwmarrin/snowflake"
	config "github.com/dynamicgo/go-config"
	_ "github.com/dynamicgo/slf4go-aliyun"
	_ "github.com/gomeshnetwork/agent/basic"
	"github.com/gomeshnetwork/gomesh"
	"github.com/gomeshnetwork/gomesh/app"
	"github.com/gomeshnetwork/tcc/engine/services/notifier"
	"github.com/gomeshnetwork/tcc/engine/services/scheduler"
	"github.com/gomeshnetwork/tcc/engine/services/storage"
	_ "github.com/lib/pq"
)

func main() {

	gomesh.LocalService("tcc.Scheduler", func(config config.Config) (gomesh.Service, error) {
		return scheduler.New(config)
	})

	gomesh.LocalService("tcc.Snowflake", func(config config.Config) (gomesh.Service, error) {
		return snowflake.NewNode(int64(config.Get("snode").Int(0)))
	})

	gomesh.LocalService("tcc.Storage", func(config config.Config) (gomesh.Service, error) {
		return storage.New(config)
	})

	gomesh.LocalService("tcc.Notifier", func(config config.Config) (gomesh.Service, error) {
		return notifier.New(config)
	})

	app.Run("tcc")
}
