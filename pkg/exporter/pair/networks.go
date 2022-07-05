package pair

type ConsensusMechanism struct {
	Name     string
	Short    string
	Priority float64
}

var (
	ProofOfWork = ConsensusMechanism{
		Name:     "Proof of Work",
		Short:    "PoW",
		Priority: 1,
	}
	ProofOfAuthority = ConsensusMechanism{
		Name:     "Proof of Authority",
		Short:    "PoA",
		Priority: 2,
	}
	ProofOfStake = ConsensusMechanism{
		Name:     "Proof of Stake",
		Short:    "PoS",
		Priority: 3,
	}
)

var DefaultConsensusMechanism = ProofOfWork

var (
	DefaultNetworkConsensusMechanism = map[uint64]ConsensusMechanism{
		5: ProofOfAuthority,
	}
)
