package slf4go

import (
	"fmt"
	"strings"
	"sync"

	config "github.com/dynamicgo/go-config"
	extend "github.com/dynamicgo/go-config-extend"
)

// LoggerBinder .
type LoggerBinder struct {
	Name    string `json:"-"` // logger name
	Backend string // logger backend
	levels  int    // logger levels
	Level   string `json:"level"`
}

func parseLevel(level string) int {
	levels := strings.Split(level, "|")

	result := 0

	for _, l := range levels {
		val, ok := Levels[l]

		if ok {
			result |= val
		}
	}

	return result
}

// Engine .
type Engine struct {
	sync.RWMutex                            // mixin rw locker
	namedBackend   map[string]LoggerFactory // Log backend
	loggerMapping  map[string]*LoggerBinder // logger backend map
	level          int                      // logger levels
	defaultBackend LoggerFactory            // default backend
	logger         map[string]Logger        // logger
}

// New .
func New() *Engine {
	return &Engine{
		namedBackend:   make(map[string]LoggerFactory),
		loggerMapping:  make(map[string]*LoggerBinder),
		logger:         make(map[string]Logger),
		level:          Trace | Debug | Info | Warn | Error | Fatal,
		defaultBackend: newColorConsole(),
	}
}

func loadBackend(config config.Config) (*LoggerBinder, LoggerFactory, error) {
	var binder LoggerBinder

	if string(config.Bytes()) == "null" {
		return nil, nil, nil
	}

	binder.Backend = "null"

	if err := config.Scan(&binder); err != nil {
		return nil, nil, err
	}

	if binder.Level == "" {
		binder.levels = Debug | Warn | Info | Error | Fatal | Trace
	} else {
		binder.levels = parseLevel(binder.Level)
	}

	if binder.Backend == "" {
		binder.Backend = "null"
	}

	backendF := register.Get(binder.Backend)

	if backendF == nil {
		return nil, nil, fmt.Errorf("unknown backend %s", binder.Backend)
	}

	backendConfig, err := extend.SubConfig(config, "config")

	if err != nil {
		return nil, nil, err
	}

	backend, err := backendF(backendConfig)

	if err != nil {
		return nil, nil, err
	}

	return &binder, backend, nil
}

// Load .
func (engine *Engine) Load(config config.Config) error {

	engine.Lock()
	defer engine.Unlock()

	defaultConfig, err := extend.SubConfig(config, "default")

	if err != nil {
		return err
	}

	binder, defaultBackend, err := loadBackend(defaultConfig)

	if err != nil {
		return err
	}

	if defaultBackend != nil {
		engine.defaultBackend = defaultBackend
		engine.level = binder.levels
	}

	loggerConfigs, err := extend.SubConfigMap(config, "logger")

	if err != nil {
		return err
	}

	for name, loggerConfig := range loggerConfigs {
		binder, backend, err := loadBackend(loggerConfig)

		if err != nil {
			return err
		}

		binder.Name = name

		engine.loggerMapping[name] = binder
		engine.namedBackend[binder.Backend] = backend
	}

	return nil
}

// Get .
func (engine *Engine) Get(name string) Logger {
	engine.RLock()
	logger, ok := engine.logger[name]

	if ok {
		engine.RUnlock()
		return logger
	}

	binder, ok := engine.loggerMapping[name]

	var backend LoggerFactory
	var level int

	if ok {
		backend, ok = engine.namedBackend[binder.Backend]

		if !ok {
			backend = engine.defaultBackend
			println("warning: unknown backend", binder.Backend)
		}

		level = binder.levels

	} else {
		backend = engine.defaultBackend
		level = engine.level
	}

	engine.RUnlock()

	engine.Lock()
	defer engine.Unlock()

	logger = &loggerWrapper{
		impl:  backend.GetLogger(name),
		level: level,
	}

	engine.logger[name] = logger

	return logger
}

var globalEngine = New()

// Backend set new slf4go backend logger factory
func Backend(factory LoggerFactory) {
	globalEngine.defaultBackend = factory
}

// Get get/create new logger by name
func Get(name string) Logger {
	return globalEngine.Get(name)
}

// SetLevel set logger level
func SetLevel(l int) {
	globalEngine.level = l
}

// GetLevel get logger level
func GetLevel() int {
	return globalEngine.level
}

// Load load loggers
func Load(config config.Config) error {
	return globalEngine.Load(config)
}
