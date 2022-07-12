package state

import (
	"time"

	"github.com/attestantio/go-eth2-client/spec"
)

// TimedBlock is a block with a timestamp.
type TimedBlock struct {
	Block  *spec.VersionedSignedBeaconBlock
	SeenAt time.Time
}
