package notifier

import (
	"fmt"
	"sync"
	"time"

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
	sync.RWMutex   // mxin rw locker
	slf4go.Logger  // logger
	agents         map[string]*agentServer
	cachesize      int
	Storage        engine.Storage `inject:"tcc.Storage"`
	reloadTimeout  time.Duration
	sessionTimeout time.Duration
}

// New .
func New(config config.Config) (engine.Notifier, error) {

	cachesize := config.Get("cached").Int(1024)

	return &notifierImpl{
		Logger:         slf4go.Get("notifier"),
		agents:         make(map[string]*agentServer),
		cachesize:      cachesize,
		reloadTimeout:  config.Get("reload").Duration(time.Minute),
		sessionTimeout: config.Get("timeout").Duration(time.Minute * 10),
	}, nil
}

func (notifier *notifierImpl) Start() error {
	go notifier.reload()
	return nil
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

	filters := make(map[string]*engine.Resource)

	for _, resource := range resources {
		filters[fmt.Sprintf("%s%s", resource.Agent, resource.Resource)] = resource
	}

	for _, resource := range filters {

		notifier.RLock()
		agent, ok := notifier.agents[resource.Agent]
		notifier.RUnlock()

		if !ok {
			notifier.WarnF("commit(%s) tx %s resource(%s,%s) to agent %s -- skipped, the agent not register",
				commit, id, resource.Require, resource.Resource, resource.Agent)

			continue
		}

		notifier.doSend(resource, agent, commit)

	}
}

func (notifier *notifierImpl) doSend(resource *engine.Resource, agent *agentServer, commit bool) {

	defer func() {
		if recover() != nil {
			notifier.ErrorF("checked closed chan for agent %s(%p) loop", agent.agent, agent)
		}
	}()

	if commit {
		agent.commit <- resource
	} else {
		agent.cancel <- resource
	}
}

func (notifier *notifierImpl) RunAgent(agent string, server tcc.Engine_AttachAgentServer) {
	as := &agentServer{
		agent:  agent,
		server: server,
		commit: make(chan *engine.Resource, notifier.cachesize),
		cancel: make(chan *engine.Resource, notifier.cachesize),
	}

	notifier.Lock()
	if old, ok := notifier.agents[agent]; ok {
		close(old.cancel)
		close(old.commit)
	}
	notifier.agents[agent] = as
	notifier.Unlock()

	notifier.doAgentLoop(as)

}

func (notifier *notifierImpl) doAgentLoop(as *agentServer) {
	for {

		notifier.InfoF("start agent %s(%p) loop", as.agent, as)

		cmd := &tcc.AgentCommandRequest{}

		var ok bool
		var resource *engine.Resource

		select {
		case resource, ok = <-as.commit:
			if !ok {
				notifier.InfoF("exit agent %s(%p) loop", as.agent, as)
				return
			}

			cmd.Command = tcc.AgentCommand_COMMMIT

		case resource, ok = <-as.cancel:
			if !ok {
				notifier.InfoF("exit agent %s(%p) loop", as.agent, as)
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

	if notifier.agents[as.agent] == as {
		delete(notifier.agents, as.agent)
		close(as.cancel)
		close(as.commit)
	}

}
