package mempool

import (
	"github.com/Qitmeer/qitmeer/log"
	"github.com/Qitmeer/qitmeer/rpc"
	"sort"
)

func (t *TxPool) API() rpc.API {
	return rpc.API{
		NameSpace: rpc.DefaultServiceNameSpace,
		Service:   NewPublicMempoolAPI(t),
		Public:    true,
	}
}

type PublicMempoolAPI struct{
	txPool *TxPool
}

func NewPublicMempoolAPI(txPool *TxPool) *PublicMempoolAPI {
	return &PublicMempoolAPI{txPool}
}

func (api *PublicMempoolAPI) GetMempool(txType *string , verbose bool) (interface{}, error){
	log.Trace("GetMempool called")
	// TODO verbose
	// The response is simply an array of the transaction hashes if the
	// verbose flag is not set.
	descs := api.txPool.TxDescs()
	hashStrings := make([]string, 0, len(descs))
	for i := range descs {
		hashStrings = append(hashStrings, descs[i].Tx.Hash().String())
	}
	sort.Strings(hashStrings)
	return hashStrings,nil
}
