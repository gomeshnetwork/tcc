package storage

import (
	"github.com/bwmarrin/snowflake"
	config "github.com/dynamicgo/go-config"
	"github.com/dynamicgo/slf4go"
	"github.com/dynamicgo/xerrors"
	"github.com/go-xorm/xorm"
	"github.com/gomeshnetwork/tcc/engine"
)

type storageImpl struct {
	slf4go.Logger
	engine *xorm.Engine    // xorm engine
	Snode  *snowflake.Node `inject:"tcc.Snowflake"`
}

// New .
func New(config config.Config) (engine.Storage, error) {

	logger := slf4go.Get("tcc.storage")

	driver := config.Get("driver").String("sqlite3")
	source := config.Get("source").String("./tcc.db")

	engine, err := xorm.NewEngine(driver, source)

	if err != nil {
		return nil, xerrors.Wrapf(err, "create xorm engine err")
	}

	return &storageImpl{
		Logger: logger,
		engine: engine,
	}, nil
}
