package types

// AmsterdamTx holds the gas limit for a single transaction in a block.
// Only the gas field is decoded; all other transaction fields are discarded.
type AmsterdamTx struct {
	Gas string `json:"gas"` // hex-encoded gas limit set by the sender
}

// AmsterdamBlock contains Amsterdam EIP-specific fields from eth_getBlockByNumber.
// Fields are hex-encoded strings as returned by the JSON-RPC API.
// Transactions contains the full transaction list (each with only Gas decoded)
// so that per-block MaxTxGas can be tracked for EIP-7825 monitoring.
type AmsterdamBlock struct {
	Number              string         `json:"number"`
	GasUsed             string         `json:"gasUsed"`
	BlockAccessListHash string         `json:"blockAccessListHash"` // EIP-7928: keccak256 of the block access list RLP
	SlotNumber          string         `json:"slotNumber"`          // EIP-7843: CL slot that produced this EL block
	Transactions        []*AmsterdamTx `json:"transactions"`        // EIP-7825: tx gas limits (full txs requested)
}

// AmsterdamLog is a single EVM log entry used to detect EIP-7708 Transfer logs.
type AmsterdamLog struct {
	Address string   `json:"address"`
	Topics  []string `json:"topics"`
}

// AmsterdamReceipt is a partial transaction receipt used for EIP-7778 gas refund delta
// and EIP-7708 Transfer log count calculations.
type AmsterdamReceipt struct {
	GasUsed string          `json:"gasUsed"`
	Logs    []*AmsterdamLog `json:"logs"`
}
