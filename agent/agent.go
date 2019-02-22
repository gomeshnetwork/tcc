package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/snowflake"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

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
func New() gomesh.TccServer {

	snode, _ := snowflake.NewNode(0)

	return &agentImpl{
		Logger:    slf4go.Get("tcc-agent"),
		resources: make(map[string]*gomesh.TccResource),
		snode:     snode,
		backoff:   config.Get("gomesh", "tcc", "backoff").Duration(time.Second * 10),
	}
}

func (agent *agentImpl) Start(config config.Config) error {

	id := config.Get("gomesh", "tcc", "id").String("")

	if id == "" {
		return xerrors.New("expect config gomesh.tcc.id")
	}

	agent.id = id

	remote := config.Get("gomesh", "tcc", "remote").String("")

	if remote == "" {
		return xerrors.New("config gomesh.tcc.remote must be set")
	}

	conn, err := grpc.Dial(remote, grpc.WithInsecure())

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
		agent.ErrorF("commit tcc session error: %s", err)
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

func (agent *agentImpl) isTccResource(grpcRequireFullMethod string) bool {
	agent.RLock()
	defer agent.RUnlock()

	_, ok := agent.resources[grpcRequireFullMethod]

	return ok
}

func (agent *agentImpl) saveResourceMetadata(ctx context.Context, rid string, localTx bool) context.Context {
	newmd := metadata.Pairs("tcc_rid_key", rid, "tcc_localtx_key", fmt.Sprintf("%v", localTx))

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	} else {
		md = md.Copy()
	}

	md = metadata.Join(md, newmd)

	return metadata.NewIncomingContext(ctx, md)
}

func (agent *agentImpl) BeforeRequire(ctx context.Context, grpcRequireFullMethod string) (context.Context, error) {

	if !agent.isTccResource(grpcRequireFullMethod) {
		agent.DebugF("[BeforeRequire] %s is not register resource api", grpcRequireFullMethod)
		return ctx, nil
	}

	rid := "R_" + agent.snode.Generate().String()

	txid, ok := gomesh.TccTxid(ctx)

	if !ok {
		session, err := gomesh.NewTcc(ctx)
		if err != nil {
			return nil, err
		}

		ctx = session.NewIncomingContext()

		txid = session.Txid()
	}

	ctx = agent.saveResourceMetadata(ctx, rid, !ok)

	agent.DebugF("[local(%v)] before tcc resource %s require with rid %s", !ok, grpcRequireFullMethod, rid)

	_, err := agent.engine.BeginLockResource(ctx, &tcc.BeginLockResourceRequest{
		Txid:     txid,
		Agent:    agent.id,
		Resource: grpcRequireFullMethod,
		Rid:      rid,
	})

	if err != nil {
		agent.ErrorF("create tcc session error: %s", err)
		return nil, err
	}

	return ctx, nil

}

func (agent *agentImpl) AfterRequire(ctx context.Context, grpcRequireFullMethod string) error {

	if !agent.isTccResource(grpcRequireFullMethod) {
		agent.DebugF("[AfterRequire] %s is not register resource api", grpcRequireFullMethod)
		return nil
	}

	txid, _ := gomesh.TccTxid(ctx)

	md, _ := metadata.FromIncomingContext(ctx)

	rid := md.Get("tcc_rid_key")[0]

	localTx := md.Get("tcc_localtx_key")[0]

	agent.DebugF("[local(%v)] after tcc resource %s require with rid %s", localTx, grpcRequireFullMethod, rid)

	_, err := agent.engine.EndLockResource(ctx, &tcc.EndLockResourceRequest{
		Txid:     txid,
		Agent:    agent.id,
		Resource: grpcRequireFullMethod,
		Rid:      rid,
	})

	if err != nil {
		agent.ErrorF("create tcc session error: %s", err)
		return err
	}

	if localTx == "true" {
		return agent.Commit(ctx, txid)
	}

	return nil
}
