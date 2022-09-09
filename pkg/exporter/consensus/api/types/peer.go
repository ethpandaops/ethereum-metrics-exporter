package types

var PeerStates = []string{
	"disconnected",
	"connected",
	"connecting",
	"disconnecting",
}

var PeerDirections = []string{
	"inbound",
	"outbound",
}

type Peer struct {
	PeerID             string `json:"peer_id"`
	ENR                string `json:"enr"`
	LastSeenP2PAddress string `json:"last_seen_p2p_address"`
	State              string `json:"state"`
	Direction          string `json:"direction"`
	Agent              string `json:"agent"`
}

func (p *Peer) DeriveAgent() Agent {
	return AgentFromString(p.Agent)
}

type Peers []Peer

type PeerCount struct {
	Disconnected  string `json:"disconnected"`
	Connected     string `json:"connected"`
	Connecting    string `json:"connecting"`
	Disconnecting string `json:"disconnecting"`
}

func (p *Peers) ByState(state string) Peers {
	var peers []Peer

	for _, peer := range *p {
		if peer.State == state {
			peers = append(peers, peer)
		}
	}

	return peers
}

func (p *Peers) ByDirection(direction string) Peers {
	var peers []Peer

	for _, peer := range *p {
		if peer.Direction == direction {
			peers = append(peers, peer)
		}
	}

	return peers
}

func (p *Peers) ByStateAndDirection(state, direction string) Peers {
	var peers []Peer

	for _, peer := range *p {
		if peer.State == state && peer.Direction == direction {
			peers = append(peers, peer)
		}
	}

	return peers
}

func (p *Peers) ByAgent(agent Agent) Peers {
	var peers []Peer

	for _, peer := range *p {
		if peer.DeriveAgent() == agent {
			peers = append(peers, peer)
		}
	}

	return peers
}

func (p *Peers) AgentCount() AgentCount {
	count := AgentCount{}

	for _, agent := range AllAgents {
		numberOfAgents := len(p.ByAgent(agent))

		switch agent {
		case AgentUnknown:
			count.Unknown = numberOfAgents
		case AgentLighthouse:
			count.Lighthouse = numberOfAgents
		case AgentNimbus:
			count.Nimbus = numberOfAgents
		case AgentTeku:
			count.Teku = numberOfAgents
		case AgentPrysm:
			count.Prysm = numberOfAgents
		case AgentLodestar:
			count.Lodestar = numberOfAgents
		}
	}

	return count
}
