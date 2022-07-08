package state

import (
	"time"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
)

// BlockTimeCalculatorBundle is a bundle of data to help with calculating proposer delay.
type BlockTimeCalculatorBundle struct {
	Genesis        *v1.Genesis
	SecondsPerSlot time.Duration
}
