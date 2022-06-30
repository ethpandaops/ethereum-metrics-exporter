package event

import (
	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
)

const (
	TopicBeaconBlock = "beacon_block"
)

type BeaconBlock struct {
	RawEvent *v1.BlockEvent
	Block    *spec.VersionedSignedBeaconBlock
}
