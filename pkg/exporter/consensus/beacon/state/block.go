package state

import (
	"time"

	"github.com/attestantio/go-eth2-client/spec"
)

type TimedBlock struct {
	Block  *spec.VersionedSignedBeaconBlock
	SeenAt time.Time
}
