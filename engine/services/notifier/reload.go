package notifier

import (
	"math/rand"
	"time"

	"github.com/gomeshnetwork/tcc"
)

func (notifier *notifierImpl) reload() {
	ticker := time.NewTicker(notifier.reloadTimeout)
	defer ticker.Stop()

	for range ticker.C {
		notifier.reloadLoop()
	}
}

func (notifier *notifierImpl) reloadLoop() {
	notifier.DebugF("start reload...")
	for _, agent := range notifier.shuffleAgent() {
		notifier.doReload(agent)
	}

	notifier.DebugF("end reload ...")
}

func (notifier *notifierImpl) doReload(agent string) {
	txs, err := notifier.Storage.QueryNotifyTx(agent)

	if err != nil {
		notifier.ErrorF("load agent %s notify tx err: %s", err)
		return
	}

	notifier.DebugF("try send notify(%d) to agent %s", len(txs), agent)

	for _, tx := range txs {
		if tx.Status == tcc.TxStatus_Confirmed {
			notifier.send(tx.ID, true)
		}

		if tx.Status == tcc.TxStatus_Canceled {
			notifier.send(tx.ID, false)
		}
	}
}

func (notifier *notifierImpl) shuffleAgent() []string {
	notifier.RLock()
	defer notifier.RUnlock()

	var agents []string

	for k := range notifier.agents {
		agents = append(agents, k)
	}

	return shuffle(agents)
}

var r = rand.New(rand.NewSource(time.Now().Unix()))

func shuffle(source []string) []string {

	// perm := r.Perm(len(source))

	// ret := make([]string, len(source))

	// for i, j := range perm {
	// 	ret[i] = source[j]
	// }

	// return ret

	for i := range source {
		j := r.Intn(i + 1)
		source[i], source[j] = source[j], source[i]
	}

	return source
}
