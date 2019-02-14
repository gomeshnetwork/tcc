package scheduler

import (
	"context"

	"github.com/dynamicgo/slf4go"

	config "github.com/dynamicgo/go-config"
	"github.com/gomeshnetwork/tcc"
	"google.golang.org/grpc"
)

type schedulerImpl struct {
	slf4go.Logger
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

func (scheduler *schedulerImpl) NewTx(context.Context, *tcc.NewTxRequest) (*tcc.NewTxResponse, error) {
	return nil, nil
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
