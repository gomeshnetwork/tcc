package storage

import (
	"github.com/bwmarrin/snowflake"
	config "github.com/dynamicgo/go-config"
	"github.com/dynamicgo/slf4go"
	"github.com/dynamicgo/xerrors"
	"github.com/dynamicgo/xxorm"
	"github.com/go-xorm/xorm"
	"github.com/gomeshnetwork/gomesh"
	"github.com/gomeshnetwork/tcc"
	"github.com/gomeshnetwork/tcc/engine"
)

type storageImpl struct {
	slf4go.Logger
	engine *xorm.Engine    // xorm engine
	Snode  *snowflake.Node `inject:"tcc.Snowflake"`
}

// New .
func New(config config.Config) (engine.Storage, error) {

	logger := slf4go.Get("tcc.storage")

	driver := config.Get("driver").String("sqlite3")
	source := config.Get("source").String("./tcc.db")

	engine, err := xorm.NewEngine(driver, source)

	if err != nil {
		return nil, xerrors.Wrapf(err, "create xorm engine err")
	}

	return &storageImpl{
		Logger: logger,
		engine: engine,
	}, nil
}

func (storage *storageImpl) NewTx(tx *engine.Transaction) error {

	_, err := storage.engine.InsertOne(tx)

	if err != nil {
		if xxorm.DuplicateKey(storage.engine, err) {
			return xerrors.Wrapf(gomesh.ErrExists, "tx %s exists", tx.ID)
		}
	}

	return nil
}

func (storage *storageImpl) UpdateTxStatus(id string, status tcc.TxStatus) (bool, error) {

	c, err := storage.engine.Where(`"i_d" = ?`, id).Cols("status").Update(&engine.Transaction{Status: status})

	if err != nil {
		return false, xerrors.Wrapf(err, "update tx %s status to %s error", id, status)
	}

	if c == 0 {
		return false, nil
	}

	return true, nil
}

func (storage *storageImpl) NewResource(resource *engine.Resource) error {
	_, err := storage.engine.InsertOne(resource)

	if err != nil {
		if xxorm.DuplicateKey(storage.engine, err) {
			return xerrors.Wrapf(gomesh.ErrExists,
				"resource(%s,%s,%s,%s) exists", resource.Tx, resource.Require, resource.Agent, resource.Resource)
		}
	}

	return nil
}

func (storage *storageImpl) UpdateResourceStatus(txid, require, agent, resource string, status tcc.TxStatus) error {
	_, err := storage.engine.
		Where(`"tx" = ? and "require" = ? and "agent" = ? and "resource" = ?`, txid, require, agent, resource).
		Cols("status").Update(&engine.Resource{Status: status})

	if err != nil {
		return xerrors.Wrapf(err, "update resource(%s,%s,%s,%s) error", txid, require, agent, resource)
	}

	return nil
}

func (storage *storageImpl) GetResourceByTx(id string) ([]*engine.Resource, error) {
	resources := make([]*engine.Resource, 0)

	err := storage.engine.Where(`"tx" = ?`, id).Find(&resources)

	if err != nil {
		return nil, xerrors.Wrapf(err, "get resources by tx %s error", id)
	}

	return resources, nil
}

func (storage *storageImpl) UpdateResourcesStatus(txid, agent, resource string, status tcc.TxStatus) error {
	_, err := storage.engine.
		Where(`"tx" = ? and "agent" = ? and "resource" = ?`, txid, agent, resource).
		Cols("status").Update(&engine.Resource{Status: status})

	if err != nil {
		return xerrors.Wrapf(err, "update resource(%s,%s,%s) error", txid, agent, resource)
	}

	return nil
}
