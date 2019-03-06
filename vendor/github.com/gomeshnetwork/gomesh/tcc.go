package gomesh

import (
	"context"
	"fmt"

	"github.com/dynamicgo/slf4go"

	"github.com/dynamicgo/xerrors"
	"google.golang.org/grpc/metadata"
)

var txidkey = "gomesh_tcc_txid"
var ridkey = "gomesh_tcc_rid"
var localkey = "gomesh_tcc_local"

// TccSession .
type TccSession interface {
	Txid() string
	Context() context.Context
	NewIncomingContext() context.Context
	Commit() error
	Cancel() error
	LocalCall(resourceName string, f func() error) error
}

type sessionImpl struct {
	slf4go.Logger
	txid      string
	tccServer TccServer
	ctx       context.Context
	inCtx     context.Context
	linked    bool
}

func (session *sessionImpl) Txid() string {
	return session.txid
}

func (session *sessionImpl) Context() context.Context {
	return session.ctx
}

func (session *sessionImpl) Commit() error {
	if !session.linked {
		return session.tccServer.Commit(session.ctx, session.txid)
	}

	return nil
}

func (session *sessionImpl) Cancel() error {
	if !session.linked {
		return session.tccServer.Cancel(session.ctx, session.txid)
	}

	return nil
}

func (session *sessionImpl) NewIncomingContext() context.Context {
	md := metadata.Pairs(txidkey, session.txid)

	return metadata.NewIncomingContext(session.ctx, md)
}

func (session *sessionImpl) LocalCall(resourceName string, f func() error) error {

	ctx := session.inCtx

	if ctx == nil {
		ctx = session.NewIncomingContext()
		session.inCtx = ctx
	}

	tccServer := GetTccServer()

	if tccServer != nil {
		var err error
		ctx, err = tccServer.BeforeRequire(ctx, resourceName)

		if err != nil {
			return xerrors.Wrapf(err, "tcc resource %s before lock err", resourceName)
		}
	}

	err := f()

	if err != nil {
		return xerrors.Wrapf(err, "tcc resource %s lock err", resourceName)
	}

	if tccServer != nil {
		err := tccServer.AfterRequire(ctx, resourceName)

		if err != nil {
			return xerrors.Wrapf(err, "tcc resource %s after lock err", resourceName)
		}
	}

	return nil
}

// NewTcc .
func NewTcc(ctx context.Context) (TccSession, error) {

	tccServer := GetTccServer()

	if tccServer == nil {
		return nil, xerrors.New("tccServer not register")
	}

	txid, ok := TccTxid(ctx)

	if !ok {
		var err error
		txid, err = tccServer.NewTx(ctx, "")

		if err != nil {
			return nil, err
		}
	}

	md := metadata.Pairs(txidkey, txid)

	session := &sessionImpl{
		Logger:    slf4go.Get("tcc"),
		txid:      txid,
		tccServer: tccServer,
		ctx:       metadata.NewOutgoingContext(ctx, md),
		linked:    ok,
	}

	return session, nil
}

// TccTxid .
func TccTxid(ctx context.Context) (string, bool) {
	return TccTxMetadata(ctx, txidkey)
}

// TccRid .
func TccRid(ctx context.Context) (string, bool) {
	return TccTxMetadata(ctx, ridkey)
}

// TccLocalTx .
func TccLocalTx(ctx context.Context) bool {
	status, ok := TccTxMetadata(ctx, localkey)

	if ok && status == "true" {
		return true
	}

	return false
}

// NewTccResourceIncomingContext .
func NewTccResourceIncomingContext(ctx context.Context, rid string, localTx bool) context.Context {
	newmd := metadata.Pairs(ridkey, rid, localkey, fmt.Sprintf("%v", localTx))

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	} else {
		md = md.Copy()
	}

	md = metadata.Join(md, newmd)

	return metadata.NewIncomingContext(ctx, md)
}

// TccTxMetadata .
func TccTxMetadata(ctx context.Context, key string) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return "", false
	}

	val := md.Get(key)

	if len(val) > 0 {
		return val[0], true
	}

	return "", false
}
