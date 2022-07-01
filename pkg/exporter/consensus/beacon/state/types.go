package state

import (
	"time"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
)

type BlockTimeCalculatorBundle struct {
	Genesis        *v1.Genesis
	SecondsPerSlot time.Duration
}
