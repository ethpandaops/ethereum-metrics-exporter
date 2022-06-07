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
