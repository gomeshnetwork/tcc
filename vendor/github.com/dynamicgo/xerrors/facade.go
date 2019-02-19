package xerrors

import (
	"fmt"
	"sync"
)

// ErrorFacade .
type ErrorFacade interface {
	Is(err, target error) bool
	As(err error, target interface{}) bool
}

var locker sync.RWMutex
var facades map[string]ErrorFacade
var once sync.Once

func initFacades() {
	facades = make(map[string]ErrorFacade)
}

// RegisterFacade .
func RegisterFacade(name string, facade ErrorFacade) {
	once.Do(initFacades)

	if _, ok := facades[name]; ok {
		panic(fmt.Sprintf("facade name register:%s", name))
	}

	facades[name] = facade
}

func loopFacade(f func(facade ErrorFacade) bool) {
	once.Do(initFacades)

	for _, facade := range facades {
		if f(facade) {
			return
		}
	}
}
