package agent

import (
	"context"
	"time"

	"github.com/dynamicgo/xerrors"
	"github.com/gomeshnetwork/tcc"
)

func (agent *agentImpl) cmdLoop(client tcc.Engine_AttachAgentClient) {
	for {
		cmd, err := client.Recv()

		if err != nil {
			agent.ErrorF("%s", xerrors.Wrapf(err, "agent recv cmd error"))
			time.Sleep(time.Second * 10)
			go agent.attach()
			return
		}

		agent.handleCmd(cmd)
	}
}

func (agent *agentImpl) handleCmd(request *tcc.AgentCommandRequest) {
	agent.RLock()
	defer agent.RUnlock()

	resource, ok := agent.resources[request.Resource]

	if !ok {
		agent.WarnF("agent %s resource %s not found ", agent.id, request.Resource)
		return
	}

	var err error
	var status tcc.TxStatus

	if request.Command == tcc.AgentCommand_COMMMIT {
		err = resource.Commit(request.Txid)
		status = tcc.TxStatus_Confirmed
	} else {
		err = resource.Cancel(request.Txid)
		status = tcc.TxStatus_Canceled
	}

	if err != nil {
		agent.ErrorF("%s", xerrors.Wrapf(err, "agent %s commit resource %s error", agent.id, request.Resource))
		return
	}

	_, err = agent.engine.ResourceStatusChanged(context.Background(), &tcc.ResourceStatusChangedRequest{
		Txid:     request.Txid,
		Resource: request.Resource,
		Status:   status,
	})

	if err != nil {
		agent.ErrorF("%s",
			xerrors.Wrapf(err, "agent %s notify resource %s status changed %s error", agent.id, request.Resource, status))
		return
	}

}

func (agent *agentImpl) attach() {

	for {
		cmd, err := agent.engine.AttachAgent(context.Background(), &tcc.AttachAgentRequest{
			Agent: agent.id,
		})

		if err != nil {
			err = xerrors.Wrapf(err, "attach agent error")
			agent.ErrorF("%s", err)
			time.Sleep(agent.backoff)
			continue
		}

		agent.DebugF("attach tcc agent -- success")

		go agent.cmdLoop(cmd)

		break
	}
}
