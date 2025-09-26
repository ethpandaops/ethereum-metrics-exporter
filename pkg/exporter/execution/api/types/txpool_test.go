package types

import (
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

func TestTXPoolStatus_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    TXPoolStatus
		wantErr bool
	}{
		{
			name: "Geth format with hex strings",
			input: `{
				"pending": "0x128",
				"queued": "0x0"
			}`,
			want: TXPoolStatus{
				Pending: hexutil.Uint64(0x128), // 296 in decimal
				Queued:  hexutil.Uint64(0x0),   // 0 in decimal
			},
			wantErr: false,
		},
		{
			name: "Nethermind format with numbers",
			input: `{
				"pending": 16,
				"queued": 0
			}`,
			want: TXPoolStatus{
				Pending: hexutil.Uint64(16),
				Queued:  hexutil.Uint64(0),
			},
			wantErr: false,
		},
		{
			name: "Mixed format",
			input: `{
				"pending": "0x10",
				"queued": 5
			}`,
			want: TXPoolStatus{
				Pending: hexutil.Uint64(0x10), // 16 in decimal
				Queued:  hexutil.Uint64(5),
			},
			wantErr: false,
		},
		{
			name: "Large hex values",
			input: `{
				"pending": "0xffffffff",
				"queued": "0x1234abcd"
			}`,
			want: TXPoolStatus{
				Pending: hexutil.Uint64(0xffffffff),
				Queued:  hexutil.Uint64(0x1234abcd),
			},
			wantErr: false,
		},
		{
			name: "Missing fields defaults to zero",
			input: `{
				"pending": "0x10"
			}`,
			want: TXPoolStatus{
				Pending: hexutil.Uint64(0x10),
				Queued:  hexutil.Uint64(0),
			},
			wantErr: false,
		},
		{
			name:    "Empty object",
			input:   `{}`,
			want:    TXPoolStatus{},
			wantErr: false,
		},
		{
			name: "Invalid hex string",
			input: `{
				"pending": "0xg123",
				"queued": "0x0"
			}`,
			wantErr: true,
		},
		{
			name: "Negative number",
			input: `{
				"pending": -1,
				"queued": 0
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got TXPoolStatus
			err := json.Unmarshal([]byte(tt.input), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("TXPoolStatus.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Pending != tt.want.Pending {
					t.Errorf("TXPoolStatus.UnmarshalJSON() Pending = %v, want %v", got.Pending, tt.want.Pending)
				}
				if got.Queued != tt.want.Queued {
					t.Errorf("TXPoolStatus.UnmarshalJSON() Queued = %v, want %v", got.Queued, tt.want.Queued)
				}
			}
		})
	}
}

func TestTXPoolStatus_UnmarshalJSON_FullResponse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    TXPoolStatus
		wantErr bool
	}{
		{
			name: "Geth full RPC response",
			input: `{
				"jsonrpc": "2.0",
				"id": 1,
				"result": {
					"pending": "0x128",
					"queued": "0x0"
				}
			}`,
			want: TXPoolStatus{
				Pending: hexutil.Uint64(0x128), // 296 in decimal
				Queued:  hexutil.Uint64(0x0),
			},
			wantErr: false,
		},
		{
			name: "Nethermind full RPC response",
			input: `{
				"jsonrpc": "2.0",
				"result": {
					"pending": 16,
					"queued": 0
				},
				"id": 1
			}`,
			want: TXPoolStatus{
				Pending: hexutil.Uint64(16),
				Queued:  hexutil.Uint64(0),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First unmarshal to get the result field
			var rpcResponse struct {
				Result json.RawMessage `json:"result"`
			}
			if err := json.Unmarshal([]byte(tt.input), &rpcResponse); err != nil {
				t.Fatalf("Failed to unmarshal RPC response: %v", err)
			}

			// Then unmarshal the result into TXPoolStatus
			var got TXPoolStatus
			err := json.Unmarshal(rpcResponse.Result, &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("TXPoolStatus.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Pending != tt.want.Pending {
					t.Errorf("TXPoolStatus.UnmarshalJSON() Pending = %v, want %v", got.Pending, tt.want.Pending)
				}
				if got.Queued != tt.want.Queued {
					t.Errorf("TXPoolStatus.UnmarshalJSON() Queued = %v, want %v", got.Queued, tt.want.Queued)
				}
			}
		})
	}
}