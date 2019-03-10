package gomesh

import (
	"context"

	config "github.com/dynamicgo/go-config"
	"github.com/dynamicgo/injector"
)

// AccessCtrl .
type AccessCtrl interface {
	Handle(ctx context.Context, method string) (context.Context, error)
	Start(config config.Config) error
}

// RegisterAccessCtrlServer .
func RegisterAccessCtrlServer(server AccessCtrl) {
	injector.Register("mesh.accessctrl", server)
}
