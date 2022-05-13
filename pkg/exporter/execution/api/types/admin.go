package types

import (
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
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
		Eth EthProtocol `json:"eth"`
	} `json:"protocols"`
}

type EthProtocol struct {
	Difficulty *big.Int    `json:"difficulty"`
	Genesis    common.Hash `json:"genesis"`
	Head       common.Hash `json:"head"`
	NetworkID  int         `json:"networkID"`
}

func (e *EthProtocol) UnmarshalJSON(data []byte) error {
	var v struct {
		Difficulty *big.Int    `json:"difficulty"`
		Genesis    common.Hash `json:"genesis"`
		Head       common.Hash `json:"head"`
		NetworkID  int         `json:"networkID"`
	}

	var objMap map[string]*json.RawMessage
	err := json.Unmarshal(data, &objMap)
	if err != nil {
		return err
	}

	var difficultyString string
	if err := json.Unmarshal(*objMap["difficulty"], &difficultyString); err != nil {
		// Its probably just an int, return the entire thing like normal
		err = json.Unmarshal(data, &v)
		if err != nil {
			return err
		}

	} else {
		// Try and parse the string back in to a big.Int
		if v.Difficulty, err = hexutil.DecodeBig(difficultyString); err != nil {
			return err
		}
	}

	e.Difficulty = v.Difficulty
	e.Genesis = v.Genesis
	e.Head = v.Head
	e.NetworkID = v.NetworkID

	return nil
}

func (n *NodeInfo) Difficulty() *big.Int {
	return n.Protocols.Eth.Difficulty
}
