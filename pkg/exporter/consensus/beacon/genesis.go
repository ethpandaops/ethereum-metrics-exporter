package beacon

import (
	"context"
	"errors"

	eth2client "github.com/attestantio/go-eth2-client"
	v1 "github.com/attestantio/go-eth2-client/api/v1"
)

func (n *node) fetchGenesis(ctx context.Context) (*v1.Genesis, error) {
	provider, isProvider := n.client.(eth2client.GenesisProvider)
	if !isProvider {
		return nil, errors.New("client does not implement eth2client.GenesisProvider")
	}

	genesis, err := provider.Genesis(ctx)
	if err != nil {
		return nil, err
	}

	n.genesis = genesis

	return genesis, nil
}
