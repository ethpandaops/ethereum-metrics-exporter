package jobs

type EventTopic string

const (
	EventTopicBlock                = "block"
	EventTopicHead                 = "head"
	EventTopicAttestation          = "attestation"
	EventTopicChainReorg           = "chain_reorg"
	EventTopicFinalizedCheckpoint  = "finalized_checkpoint"
	EventTopicVoluntaryExit        = "voluntary_exit"
	EventTopicContributionAndProof = "contribution_and_proof"
)
