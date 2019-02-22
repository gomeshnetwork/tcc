package gomesh

import (
	"context"
	"fmt"

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
}

type sessionImpl struct {
	txid      string
	tccServer TccServer
	ctx       context.Context
}

func (session *sessionImpl) Txid() string {
	return session.txid
}

func (session *sessionImpl) Context() context.Context {
	return session.ctx
}

func (session *sessionImpl) Commit() error {
	return session.tccServer.Commit(session.ctx, session.txid)
}

func (session *sessionImpl) Cancel() error {
	return session.tccServer.Cancel(session.ctx, session.txid)
}

func (session *sessionImpl) NewIncomingContext() context.Context {
	md := metadata.Pairs(txidkey, session.txid)

	return metadata.NewIncomingContext(session.ctx, md)
}

// NewTcc .
func NewTcc(ctx context.Context) (TccSession, error) {

	tccServer := GetTccServer()

	if tccServer == nil {
		return nil, xerrors.New("tccServer not register")
	}

	parentTxid, _ := TccTxid(ctx)

	txid, err := tccServer.NewTx(ctx, parentTxid)

	if err != nil {
		return nil, err
	}

	md := metadata.Pairs(txidkey, txid)

	session := &sessionImpl{
		txid:      txid,
		tccServer: tccServer,
		ctx:       metadata.NewOutgoingContext(ctx, md),
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
	status, ok := TccTxMetadata(ctx, ridkey)

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
