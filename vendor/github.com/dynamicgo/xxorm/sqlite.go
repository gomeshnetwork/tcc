package xxorm

import (
	"github.com/go-xorm/core"
	sqlite3 "github.com/mattn/go-sqlite3"
)

type sqliteDialect struct {
}

func (dialect *sqliteDialect) DuplicateKey(err error) bool {
	if v, ok := err.(sqlite3.Error); ok {
		if v.ExtendedCode == 2067 {
			return true
		}
		return false
	}

	return false
}

func init() {
	RegisterDialect(core.SQLITE, &sqliteDialect{})
}
