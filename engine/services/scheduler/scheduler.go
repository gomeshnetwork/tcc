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
	SNode   *snowflake.Node `inject:"tcc.Snowflake"` // inject snowflake node
	Storage engine.Storage  `inject:"tcc.Storage"`   // inject storage service
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

	tx := &engine.Transaction{
		ID:     scheduler.SNode.Generate().String(),
		PID:    request.Txid,
		Status: tcc.TxStatus_Created,
	}

	if err := scheduler.Storage.NewTx(tx); err != nil {
		return nil, err
	}

	return &tcc.NewTxResponse{
		Txid: tx.ID,
	}, nil
}

func (scheduler *schedulerImpl) Commit(context.Context, *tcc.CommitTxRequest) (*tcc.CommitTxResponse, error) {

	return nil, nil
}

func (scheduler *schedulerImpl) Cancel(context.Context, *tcc.CancelTxRequest) (*tcc.CancelTxResponse, error) {
	return nil, nil
}

func (scheduler *schedulerImpl) BeforeRequire(context.Context, *tcc.BeforeRequireRequest) (*tcc.BeforeRequireRespose, error) {
	return nil, nil
}

func (scheduler *schedulerImpl) AfterRequire(context.Context, *tcc.AfterRequireRequest) (*tcc.AfterRequireRespose, error) {
	return nil, nil
}

func (scheduler *schedulerImpl) AttachAgent(*tcc.AttachAgentRequest, tcc.Engine_AttachAgentServer) error {
	return nil
}
