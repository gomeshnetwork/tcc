package scheduler

import (
	"context"

	"github.com/bwmarrin/snowflake"
	"github.com/gomeshnetwork/tcc/engine"

	"github.com/dynamicgo/slf4go"

	config "github.com/dynamicgo/go-config"
	"github.com/gomeshnetwork/tcc"
	"google.golang.org/grpc"
)

type schedulerImpl struct {
	slf4go.Logger
	SNode    *snowflake.Node `inject:"tcc.Snowflake"` // inject snowflake node
	Storage  engine.Storage  `inject:"tcc.Storage"`   // inject storage service
	Notifier engine.Notifier `inject:"tcc.Notifier"`  // inject resource manager notifier
}

// New .
func New(config config.Config) (tcc.EngineServer, error) {
	return &schedulerImpl{
		Logger: slf4go.Get("tcc-scheduler"),
	}, nil
}

func (scheduler *schedulerImpl) GrpcHandle(server *grpc.Server) error {
	tcc.RegisterEngineServer(server, scheduler)
	scheduler.InfoF("register grpc server for tcc.Scheduler ")
	return nil
}

func (scheduler *schedulerImpl) NewTx(ctx context.Context, request *tcc.NewTxRequest) (*tcc.NewTxResponse, error) {

	txid, err := scheduler.newTx(ctx, request.Txid)

	return &tcc.NewTxResponse{
		Txid: txid,
	}, err
}

func (scheduler *schedulerImpl) newTx(ctx context.Context, pid string) (string, error) {
	tx := &engine.Transaction{
		ID:     scheduler.SNode.Generate().String(),
		PID:    pid,
		Status: tcc.TxStatus_Created,
	}

	if err := scheduler.Storage.NewTx(tx); err != nil {
		return "", err
	}

	scheduler.DebugF("new tx %s", tx.ID)

	return tx.ID, nil
}

func (scheduler *schedulerImpl) Commit(ctx context.Context, request *tcc.CommitTxRequest) (*tcc.CommitTxResponse, error) {

	ok, err := scheduler.Storage.UpdateTxStatus(request.Txid, tcc.TxStatus_Confirmed)

	if err != nil {
		return nil, err
	}

	if ok {
		scheduler.Notifier.CommitTx(request.Txid)
	}

	return &tcc.CommitTxResponse{}, nil
}

func (scheduler *schedulerImpl) Cancel(ctx context.Context, request *tcc.CancelTxRequest) (*tcc.CancelTxResponse, error) {
	ok, err := scheduler.Storage.UpdateTxStatus(request.Txid, tcc.TxStatus_Canceled)

	if err != nil {
		return nil, err
	}

	if ok {
		scheduler.Notifier.CancelTx(request.Txid)
	}

	return &tcc.CancelTxResponse{}, nil
}

func (scheduler *schedulerImpl) BeginLockResource(ctx context.Context, request *tcc.BeginLockResourceRequest) (*tcc.BeginLockResourceRespose, error) {

	resource := &engine.Resource{
		ID:       "R_" + scheduler.SNode.Generate().String(),
		Tx:       request.Txid,
		Require:  request.Rid,
		Agent:    request.Agent,
		Resource: request.Resource,
		Status:   tcc.TxStatus_Created,
	}

	if err := scheduler.Storage.NewResource(resource); err != nil {
		return nil, err
	}

	return &tcc.BeginLockResourceRespose{}, nil
}

func (scheduler *schedulerImpl) EndLockResource(ctx context.Context, request *tcc.EndLockResourceRequest) (*tcc.EndLockResourceRespose, error) {

	if err := scheduler.Storage.
		UpdateResourceStatus(request.Txid, request.Rid, request.Agent, request.Resource, tcc.TxStatus_Locked); err != nil {
		return nil, err
	}

	return &tcc.EndLockResourceRespose{}, nil
}

func (scheduler *schedulerImpl) AttachAgent(request *tcc.AttachAgentRequest, agentServer tcc.Engine_AttachAgentServer) error {
	scheduler.Notifier.RunAgent(request.Agent, agentServer)
	return nil
}

func (scheduler *schedulerImpl) ResourceStatusChanged(ctx context.Context, request *tcc.ResourceStatusChangedRequest) (*tcc.ResourceStatusChangedRespose, error) {
	if err := scheduler.Storage.UpdateResourcesStatus(request.Txid, request.Agent, request.Resource, request.Status); err != nil {
		return nil, err
	}

	return &tcc.ResourceStatusChangedRespose{}, nil
}
