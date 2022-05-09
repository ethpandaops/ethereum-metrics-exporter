package execution

type SyncStatus struct {
	IsSyncing     bool
	StartingBlock uint64
	CurrentBlock  uint64
	HighestBlock  uint64
}

func (e *SyncStatus) Percent() float64 {
	if !e.IsSyncing {
		return 100
	}

	return float64(e.CurrentBlock) / float64(e.HighestBlock) * 100
}
