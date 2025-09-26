package types

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

// TXPoolStatus is the information about the transaction pool.
type TXPoolStatus struct {
	Pending hexutil.Uint64 `json:"pending"`
	Queued  hexutil.Uint64 `json:"queued"`
}

// UnmarshalJSON implements custom unmarshaling to handle both hex (Geth) and uint (Nethermind) formats
func (t *TXPoolStatus) UnmarshalJSON(data []byte) error {
	// Try to unmarshal into a map first to handle both formats
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Helper function to parse either hex string or number
	parseValue := func(key string) (hexutil.Uint64, error) {
		val, ok := raw[key]
		if !ok {
			return 0, nil // Field not present, use zero value
		}

		switch v := val.(type) {
		case string:
			// Handle hex string (Geth format)
			if strings.HasPrefix(v, "0x") || strings.HasPrefix(v, "0X") {
				decoded, err := hexutil.DecodeUint64(v)
				if err != nil {
					return 0, fmt.Errorf("failed to decode hex value for %s: %w", key, err)
				}

				return hexutil.Uint64(decoded), nil
			}
			// Try parsing as decimal string
			n, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("failed to parse string value for %s: %w", key, err)
			}

			return hexutil.Uint64(n), nil
		case float64:
			// Handle numeric value (Nethermind format)
			if v < 0 || v > float64(^uint64(0)) {
				return 0, fmt.Errorf("value for %s out of uint64 range: %v", key, v)
			}

			return hexutil.Uint64(uint64(v)), nil
		default:
			return 0, fmt.Errorf("unexpected type for %s: %T", key, v)
		}
	}

	// Parse pending and queued values
	pending, err := parseValue("pending")
	if err != nil {
		return err
	}

	t.Pending = pending

	queued, err := parseValue("queued")
	if err != nil {
		return err
	}

	t.Queued = queued

	return nil
}
