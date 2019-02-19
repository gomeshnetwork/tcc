package sync

import (
	"fmt"
	"sync"

	"github.com/dynamicgo/xerrors"

	config "github.com/dynamicgo/go-config"
	"github.com/dynamicgo/slf4go"
	"github.com/go-xorm/xorm"
)

// Handler sync handler prototype
type Handler func() []interface{}

type syncRegisterImpl struct {
	slf4go.Logger
	sync.RWMutex
	handlers map[string]Handler
}

func (register *syncRegisterImpl) Register(name string, handler Handler) {
	register.Lock()
	defer register.Unlock()

	if _, ok := register.handlers[name]; ok {
		panic(fmt.Errorf("duplicate db sync handler: %s", name))
	}

	register.handlers[name] = handler
}

func (register *syncRegisterImpl) Sync(name string, engine *xorm.Engine) error {
	register.RLock()
	defer register.RUnlock()

	handler, ok := register.handlers[name]

	if !ok {
		return nil
	}

	return engine.Sync2(handler()...)
}

var register = &syncRegisterImpl{
	Logger:   slf4go.Get("orm"),
	handlers: make(map[string]Handler),
}

// Register .
func Register(name string, handler Handler) {
	register.DebugF("register orm module %s", name)
	register.Register(name, handler)
}

// Sync .
func Sync(name string, engine *xorm.Engine) error {
	return register.Sync(name, engine)
}

// WithConfig .
func WithConfig(config config.Config) error {
	for name, handlers := range register.handlers {
		register.DebugF("load db: %s", name)

		db, err := loadDB(config, name)

		if err != nil {
			return xerrors.Wrapf(err, "load db %s error", name)
		}

		register.DebugF("sync db: %s", name)

		err = db.Sync2(handlers()...)

		db.Close()

		if err != nil {
			return xerrors.Wrapf(err, "db.Sync2 error")
		}

	}

	return nil
}

func loadDB(config config.Config, name string) (*xorm.Engine, error) {
	driver := config.Get("database", name, "driver").String("driver")
	source := config.Get("database", name, "source").String("source")

	return xorm.NewEngine(driver, source)
}
