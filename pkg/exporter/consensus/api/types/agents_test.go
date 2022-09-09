package types

import "testing"

func TestAgentParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input  string
		expect Agent
	}{
		{"Prysm/v2.0.2/4a4a7e97dfd2285a5e48a178f693d870e9a4ff60", AgentPrysm},
		{"Lighthouse/v3.1.0-aa022f4/x86_64-linux", AgentLighthouse},
		{"nimbus", AgentNimbus},
		{"teku/teku/v22.9.0/linux-x86_64/-privatebuild-openjdk64bitservervm-java-17", AgentTeku},
		{"Lodestar/v0.32.0-rc.0-1-gc3b5b6a9/linux-x64/nodejs", AgentLodestar},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			t.Parallel()
			if actual := AgentFromString(test.input); actual != test.expect {
				t.Errorf("Expected %s, got %s", test.expect, actual)
			}
		})
	}
}
