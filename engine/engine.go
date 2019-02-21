package engine

import (
	"time"

	"github.com/gomeshnetwork/tcc"
)

// Transaction .
type Transaction struct {
	ID          string       `xorm:"pk"`      // txid
	PID         string       `xorm:"index"`   // parent txid
	Status      tcc.TxStatus `xorm:"index"`   // transaction status
	CreatedTime time.Time    `xorm:"created"` // create time
	UpdatedTime time.Time    `xorm:"updated"` // updated time
}

// TableName .
func (table *Transaction) TableName() string {
	return "tcc_engine_transaction"
}

// Resource tcc resource status table
type Resource struct {
	ID          string       `xorm:"pk"`                       // txid
	Tx          string       `xorm:"unique(tx_req_agent_res)"` // resource bind transaction
	Require     string       `xorm:"unique(tx_req_agent_res)"` // resource require id
	Agent       string       `xorm:"unique(tx_req_agent_res)"` // resource require agent id
	Resource    string       `xorm:"unique(tx_req_agent_res)"` // resource require agent id
	Status      tcc.TxStatus `xorm:"index"`                    // transaction status
	CreatedTime time.Time    `xorm:"created"`                  // create time
	UpdatedTime time.Time    `xorm:"updated"`                  // updated time
}

// TableName .
func (table *Resource) TableName() string {
	return "tcc_engine_resource"
}

// Storage .
type Storage interface {
	NewTx(tx *Transaction) error
	UpdateTxStatus(id string, status tcc.TxStatus) (bool, error)
	NewResource(resource *Resource) error
	UpdateResourceStatus(txid, agent, resource, rid string, status tcc.TxStatus) error
	UpdateResourcesStatus(txid, agent, resource string, status tcc.TxStatus) error
	GetResourceByTx(id string) ([]*Resource, error)
}

// Notifier .
type Notifier interface {
	CommitTx(id string)
	CancelTx(id string)
	RunAgent(agent string, server tcc.Engine_AttachAgentServer)
}
