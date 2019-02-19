package xxorm

import (
	"fmt"
	"sync"
)

// Dialect .
type Dialect interface {
	DuplicateKey(err error) bool
}

type registerImpl struct {
	sync.RWMutex
	dialects map[string]Dialect
}

var dialects *registerImpl
var once sync.Once

func initOnce() {
	dialects = &registerImpl{
		dialects: make(map[string]Dialect),
	}
}

// RegisterDialect  .
func RegisterDialect(name string, dialect Dialect) {
	once.Do(initOnce)

	dialects.Lock()
	defer dialects.Unlock()

	dialects.dialects[name] = dialect
}

func getDialect(name string) Dialect {
	dialects.RLock()
	defer dialects.RUnlock()

	dialect := dialects.dialects[name]

	if dialect == nil {
		panic(fmt.Errorf("unknown xorm errors dialect for database %s", name))
	}

	return dialect
}
