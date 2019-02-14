package main

import (
	"github.com/dynamicgo/xxorm/sync"
	"github.com/gomeshnetwork/tcc/engine"
	_ "github.com/lib/pq"
)

func main() {
	sync.Register("tcc", func() []interface{} {
		return []interface{}{
			new(engine.Transaction),
			new(engine.Resource),
		}
	})

	sync.Run("tcc.syncdb")
}
