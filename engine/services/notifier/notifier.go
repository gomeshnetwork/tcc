package notifier

import (
	"fmt"
	"sync"

	config "github.com/dynamicgo/go-config"
	"github.com/dynamicgo/slf4go"
	"github.com/dynamicgo/xerrors"
	"github.com/gomeshnetwork/tcc"
	"github.com/gomeshnetwork/tcc/engine"
)

type agentServer struct {
	agent  string
	server tcc.Engine_AttachAgentServer
	commit chan *engine.Resource
	cancel chan *engine.Resource
}

type notifierImpl struct {
	sync.RWMutex  // mxin rw locker
	slf4go.Logger // logger
	agents        map[string]*agentServer
	cachesize     int
	Storage       engine.Storage `inject:"tcc.Storage"`
}

// New .
func New(config config.Config) (engine.Notifier, error) {

	cachesize := config.Get("cached").Int(1024)

	return &notifierImpl{
		Logger:    slf4go.Get("notifier"),
		agents:    make(map[string]*agentServer),
		cachesize: cachesize,
	}, nil
}

func (notifier *notifierImpl) CommitTx(id string) {
	notifier.send(id, true)
}

func (notifier *notifierImpl) CancelTx(id string) {
	notifier.send(id, false)
}

func (notifier *notifierImpl) send(id string, commit bool) {
	resources, err := notifier.Storage.GetResourceByTx(id)

	if err != nil {
		err = xerrors.Wrapf(err, "commit tx %s error", id)
		notifier.ErrorF("%s", err)
		return
	}

	if len(resources) == 0 {
		notifier.InfoF("commit tx %s -- no resources locked", id)
		return
	}

	notifier.RLock()
	defer notifier.RUnlock()

	filters := make(map[string]*engine.Resource)

	for _, resource := range resources {
		filters[fmt.Sprintf("%s%s", resource.Agent, resource.Resource)] = resource
	}

	for _, resource := range filters {
		agent, ok := notifier.agents[resource.Agent]

		if !ok {
			notifier.WarnF("commit(%s) tx %s resource(%s,%s) to agent %s -- skipped, the agent not register",
				commit, id, resource.Require, resource.Resource, resource.Agent)

			continue
		}

		if commit {
			agent.commit <- resource
		} else {
			agent.cancel <- resource
		}
	}
}

func (notifier *notifierImpl) Register(agent string, server tcc.Engine_AttachAgentServer) {
	as := &agentServer{
		agent:  agent,
		server: server,
		commit: make(chan *engine.Resource, notifier.cachesize),
		cancel: make(chan *engine.Resource, notifier.cachesize),
	}

	go notifier.doAgentLoop(as)

	notifier.Lock()
	defer notifier.Unlock()

	notifier.agents[agent] = as
}

func (notifier *notifierImpl) doAgentLoop(as *agentServer) {
	for {

		cmd := &tcc.AgentCommandRequest{}

		var ok bool
		var resource *engine.Resource

		select {
		case resource, ok = <-as.commit:
			if !ok {
				return
			}

			cmd.Command = tcc.AgentCommand_COMMMIT

		case resource, ok = <-as.cancel:
			if !ok {
				return
			}

			cmd.Command = tcc.AgentCommand_Cancel
		}

		cmd.Resource = resource.Resource
		cmd.Txid = resource.Tx

		notifier.InfoF("send agent command to %s: %s", as.agent, cmd)

		if err := as.server.Send(cmd); err != nil {
			notifier.ErrorF("cmd to agent %s err: %s", as.agent, err)
			notifier.closeAgentServer(as)
		}
	}
}

func (notifier *notifierImpl) closeAgentServer(as *agentServer) {
	notifier.Lock()
	defer notifier.Unlock()

	delete(notifier.agents, as.agent)
	close(as.cancel)
	close(as.commit)

}
