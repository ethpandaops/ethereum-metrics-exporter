package types

// AmsterdamBlock contains Amsterdam EIP-specific fields from eth_getBlockByNumber.
// Fields are hex-encoded strings as returned by the JSON-RPC API.
type AmsterdamBlock struct {
	Number              string `json:"number"`
	GasUsed             string `json:"gasUsed"`
	BlockAccessListHash string `json:"blockAccessListHash"` // EIP-7928: keccak256 of the block access list RLP
	SlotNumber          string `json:"slotNumber"`          // EIP-7843: CL slot that produced this EL block
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
