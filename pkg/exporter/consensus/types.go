package consensus

type SyncStatus struct {
	IsSyncing         bool
	HeadSlot          uint64
	SyncDistance      uint64
	EstimatedHeadSlot uint64
}

func (c *SyncStatus) Percent() float64 {
	if !c.IsSyncing {
		return 100
	}

	return (float64(c.HeadSlot) / float64(c.EstimatedHeadSlot) * 100)
}

type BlockchainSlots struct {
	Head      uint64
	Genesis   uint64
	Finalized uint64
}

type Fork struct {
	Name  string
	Epoch uint64
}
