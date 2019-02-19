package slf4go

import (
	"sync"

	config "github.com/dynamicgo/go-config"
)

// BackendF backend factory function
type BackendF func(config config.Config) (LoggerFactory, error)

type backendRegister struct {
	sync.RWMutex
	handlers map[string]BackendF
}

// Register .
type Register interface {
	Set(backend string, backendF BackendF)
	Get(backend string) BackendF
}

func newBackendRegister() Register {
	return &backendRegister{
		handlers: make(map[string]BackendF),
	}
}

func (register *backendRegister) Set(backend string, backendF BackendF) {
	register.Lock()
	defer register.Unlock()
	register.handlers[backend] = backendF
}

func (register *backendRegister) Get(backend string) BackendF {
	register.RLock()
	defer register.RUnlock()

	return register.handlers[backend]
}

var register = newBackendRegister()

// RegisterBackend register backend
func RegisterBackend(backend string, backendF BackendF) {
	register.Set(backend, backendF)
}
