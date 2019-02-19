package xxorm

import (
	"github.com/go-xorm/xorm"
)

// PageOrder type
type PageOrder string

// Order types
var (
	ASC  = PageOrder("asc")
	DESC = PageOrder("desc")
)

// MaxPageSize .
var MaxPageSize = uint64(200)

// Page .
type Page struct {
	Offset  uint64    `binding:"required" json:"offset"`
	Size    uint64    `binding:"required" json:"size"`
	OrderBy string    `json:"orderby"`
	Order   PageOrder `json:"order"`
}

// DuplicateKey check if error is a duplicate key error
func DuplicateKey(db *xorm.Engine, err error) bool {

	dialect := getDialect(string(db.Dialect().DBType()))

	return dialect.DuplicateKey(err)
}

// Paged .
func Paged(session *xorm.Session, page Page) *xorm.Session {
	if page.Size > MaxPageSize {
		page.Size = MaxPageSize
	}

	session = session.Limit(int(page.Size), int(page.Offset))

	if page.OrderBy != "" {
		if page.Order == DESC {
			session = session.Desc(page.OrderBy)
		} else {
			session = session.Asc(page.OrderBy)
		}
	}

	return session
}
