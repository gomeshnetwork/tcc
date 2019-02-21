package gomesh

import (
	"github.com/gomeshnetwork/gomesh"
	"github.com/gomeshnetwork/tcc/agent"
)

func init() {
	gomesh.RegisterTccServer(agent.New())
}
