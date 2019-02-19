package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/snowflake"

	"google.golang.org/grpc"

	"github.com/dynamicgo/xerrors"

	"github.com/dynamicgo/slf4go"
	"github.com/gomeshnetwork/tcc"

	config "github.com/dynamicgo/go-config"
	"github.com/gomeshnetwork/gomesh"
)

type agentImpl struct {
	sync.RWMutex                                 // mixin mutex
	slf4go.Logger                                // mixin logger
	id            string                         // agent id
	engine        tcc.EngineClient               // engine client
	resources     map[string]*gomesh.TccResource // register local resources
	snode         *snowflake.Node                // snode
	backoff       time.Duration                  // attach backoff time

}

// New create new agent which implement gomesh.TccServer interface
func New(config config.Config) (gomesh.TccServer, error) {

	snode, err := snowflake.NewNode(0)

	if err != nil {
		return nil, xerrors.Wrapf(err, "create snode error")
	}

	id := config.Get("gomesh", "tcc", "id").String("")

	if id == "" {
		return nil, xerrors.New("expect config gomesh.tcc.id")
	}

	return &agentImpl{
		Logger:    slf4go.Get("tcc-agent"),
		resources: make(map[string]*gomesh.TccResource),
		snode:     snode,
		id:        id,
		backoff:   config.Get("gomesh", "tcc", "backoff").Duration(time.Second * 10),
	}, nil
}

func (agent *agentImpl) Start(config config.Config) error {

	remote := config.Get("gomesh", "tcc", "remote").String("")

	if remote == "" {
		return xerrors.New("config gomesh.tcc.remote must be set")
	}

	conn, err := grpc.Dial(remote)

	if err != nil {
		return xerrors.Wrapf(err, "grpc connect to %s error", remote)
	}

	agent.engine = tcc.NewEngineClient(conn)

	go agent.attach()

	return nil
}

func (agent *agentImpl) Register(tccResource gomesh.TccResource) error {

	agent.Lock()
	defer agent.Unlock()

	_, ok := agent.resources[tccResource.GrpcRequireFullMethod]

	if ok {
		return xerrors.New(fmt.Sprintf("resource exits: %s", tccResource.GrpcRequireFullMethod))
	}

	agent.resources[tccResource.GrpcRequireFullMethod] = &tccResource

	return nil
}

func (agent *agentImpl) NewTx(ctx context.Context, parentTxid string) (string, error) {

	resp, err := agent.engine.NewTx(ctx, &tcc.NewTxRequest{
		Txid: parentTxid,
	})

	if err != nil {
		agent.ErrorF("create tcc session error: %s", err)
		return "", err
	}

	return resp.Txid, nil
}

func (agent *agentImpl) Commit(ctx context.Context, txid string) error {
	_, err := agent.engine.Commit(ctx, &tcc.CommitTxRequest{
		Txid: txid,
	})

	if err != nil {
		agent.ErrorF("create tcc session error: %s", err)
		return err
	}

	return nil
}

func (agent *agentImpl) Cancel(ctx context.Context, txid string) error {
	_, err := agent.engine.Cancel(ctx, &tcc.CancelTxRequest{
		Txid: txid,
	})

	if err != nil {
		agent.ErrorF("create tcc session error: %s", err)
		return err
	}

	return nil
}

func (agent *agentImpl) BeforeRequire(ctx context.Context, txid string, grpcRequireFullMethod string) (string, error) {

	key := "R_" + agent.snode.Generate().String()

	_, err := agent.engine.BeginLockResource(ctx, &tcc.BeginLockResourceRequest{
		Txid:     txid,
		Agent:    agent.id,
		Resource: grpcRequireFullMethod,
		Rid:      key,
	})

	if err != nil {
		agent.ErrorF("create tcc session error: %s", err)
		return "", err
	}

	return key, nil

}

func (agent *agentImpl) AfterRequire(ctx context.Context, txid string, grpcRequireFullMethod string, key string) error {
	_, err := agent.engine.EndLockResource(ctx, &tcc.EndLockResourceRequest{
		Txid:     txid,
		Agent:    agent.id,
		Resource: grpcRequireFullMethod,
		Rid:      key,
	})

	if err != nil {
		agent.ErrorF("create tcc session error: %s", err)
		return err
	}

	return nil
}
