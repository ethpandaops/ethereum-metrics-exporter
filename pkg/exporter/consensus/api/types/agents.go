package types

import (
	"strings"
)

type Agent string

const (
	AgentUnknown    Agent = "unknown"
	AgentLighthouse Agent = "lighthouse"
	AgentNimbus     Agent = "nimbus"
	AgentTeku       Agent = "teku"
	AgentPrysm      Agent = "prysm"
	AgentLodestar   Agent = "lodestar"
)

var AllAgents = []Agent{
	AgentUnknown,
	AgentLighthouse,
	AgentNimbus,
	AgentTeku,
	AgentPrysm,
	AgentLodestar,
}

type AgentCount struct {
	Unknown    int `json:"unknown"`
	Lighthouse int `json:"lighthouse"`
	Nimbus     int `json:"nimbus"`
	Teku       int `json:"teku"`
	Prysm      int `json:"prysm"`
	Lodestar   int `json:"lodestar"`
}

func AgentFromString(agent string) Agent {
	asLower := strings.ToLower(agent)

	if strings.Contains(asLower, "lighthouse") {
		return AgentLighthouse
	}

	if strings.Contains(asLower, "nimbus") {
		return AgentNimbus
	}

	if strings.Contains(asLower, "teku") {
		return AgentTeku
	}

	if strings.Contains(asLower, "prysm") {
		return AgentPrysm
	}

	if strings.Contains(asLower, "lodestar") {
		return AgentLodestar
	}

	return AgentUnknown
}
