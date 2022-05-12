package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type NodeInfo struct {
	Enode      string `json:"enode"`
	ID         string `json:"id"`
	IP         string `json:"ip"`
	ListenAddr string `json:"listenAddr"`
	Name       string `json:"name"`
	Ports      struct {
		Discovery int `json:"discovery"`
		Listener  int `json:"listener"`
	} `json:"ports"`
	Protocols struct {
		Eth struct {
			Difficulty *big.Int    `json:"difficulty"`
			Genesis    common.Hash `json:"genesis"`
			Head       common.Hash `json:"head"`
			NetworkID  int         `json:"networkID"`
		} `json:"eth"`
	} `json:"protocols"`
}
