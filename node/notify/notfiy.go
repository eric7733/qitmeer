package notify

import (
	"github.com/Qitmeer/qitmeer-lib/core/types"
	"github.com/Qitmeer/qitmeer/core/message"
)

// Notify interface manage message announce & relay & notification between mempool, websocket, gbt long pull
// and rpc server.
type Notify interface {
	AnnounceNewTransactions(newTxs []*types.Tx)
	RelayInventory(invVect *message.InvVect, data interface{})
	BroadcastMessage(msg message.Message)
}
