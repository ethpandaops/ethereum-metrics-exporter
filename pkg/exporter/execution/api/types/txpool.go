package types

import "github.com/ethereum/go-ethereum/common/hexutil"

type TXPoolStatus struct {
	Pending hexutil.Uint64 `json:"pending"`
	Queued  hexutil.Uint64 `json:"queued"`
}
