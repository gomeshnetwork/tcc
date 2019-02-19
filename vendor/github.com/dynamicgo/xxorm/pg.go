package xxorm

import (
	"github.com/go-xorm/core"
	"github.com/lib/pq"
)

type pgDialect struct {
}

func (dialect *pgDialect) DuplicateKey(err error) bool {
	if v, ok := err.(*pq.Error); ok {
		if v.Code == "23505" {
			return true
		}
		return false
	}

	return false
}

func init() {
	RegisterDialect(core.POSTGRES, &pgDialect{})
}
