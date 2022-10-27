package pair

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethpandaops/ethereum-metrics-exporter/pkg/exporter/consensus/beacon"
	"github.com/onrik/ethrpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// Metrics reports pair metrics
type Metrics interface {
	// StartAsync starts the disk usage metrics collection.
	StartAsync(ctx context.Context)
}

type pair struct {
	log logrus.FieldLogger

	consensusMechanism *prometheus.GaugeVec

	executionClient *ethclient.Client
	beacon          beacon.Node
	ethrpcClient    *ethrpc.EthRPC
	bootstrapped    bool
	executionURL    string

	totalDifficulty *big.Int
	networkID       *big.Int

	networkIDFetchedAt time.Time
	tdFetchedAt        time.Time
}

// NewMetrics returns a new Metrics instance.
func NewMetrics(ctx context.Context, log logrus.FieldLogger, namespace string, beac beacon.Node, executionURL string) (Metrics, error) {
	p := &pair{
		log: log,

		executionURL: executionURL,

		beacon: beac,

		executionClient: nil,
		ethrpcClient:    nil,

		bootstrapped: false,

		consensusMechanism: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "consensus_mechanism",
				Help:      "Consensus mechanism used",
			},
			[]string{
				"consensus_mechanism",
				"consensus_mechanism_short",
			},
		),

		totalDifficulty: big.NewInt(0),
	}

	prometheus.MustRegister(p.consensusMechanism)

	return p, nil
}

func (p *pair) Bootstrap(ctx context.Context) error {
	executionClient, err := ethclient.Dial(p.executionURL)
	if err != nil {
		return err
	}

	p.executionClient = executionClient

	p.ethrpcClient = ethrpc.New(p.executionURL)

	p.bootstrapped = true

	return nil
}

func (p *pair) StartAsync(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second * 5):
				if !p.bootstrapped {
					if err := p.Bootstrap(ctx); err != nil {
						continue
					}
				}

				if time.Since(p.tdFetchedAt) > 12*time.Second {
					if err := p.fetchTotalDifficulty(ctx); err != nil {
						p.log.WithError(err).Error("Failed to fetch total difficulty")
					}
				}

				if time.Since(p.networkIDFetchedAt) > 15*time.Minute {
					if err := p.fetchNetworkID(ctx); err != nil {
						p.log.WithError(err).Error("Failed to fetch network ID")
					}
				}

				if err := p.deriveConsensusMechanism(ctx); err != nil {
					p.log.WithError(err).Error("Failed to derive consensus mechanism")
				}
			}
		}
	}()
}

func (p *pair) fetchTotalDifficulty(ctx context.Context) error {
	mostRecentBlockNumber, err := p.executionClient.BlockNumber(ctx)
	if err != nil {
		return err
	}

	block, err := p.ethrpcClient.EthGetBlockByNumber(int(mostRecentBlockNumber), false)
	if err != nil {
		return err
	}

	if block == nil {
		return errors.New("empty block found")
	}

	p.totalDifficulty = &block.TotalDifficulty

	p.tdFetchedAt = time.Now()

	return nil
}

func (p *pair) fetchNetworkID(ctx context.Context) error {
	networkID, err := p.executionClient.NetworkID(ctx)
	if err != nil {
		return err
	}

	p.networkID = networkID

	p.networkIDFetchedAt = time.Now()

	return nil
}

func (p *pair) deriveConsensusMechanism(ctx context.Context) error {
	spec, err := p.beacon.GetSpec(ctx)
	if err != nil {
		return err
	}

	if p.totalDifficulty == big.NewInt(0) {
		return errors.New("total difficulty not fetched")
	}

	if p.networkID == big.NewInt(0) {
		return errors.New("network ID not fetched")
	}

	consensusMechanism := DefaultConsensusMechanism

	networkID := uint64(1)

	if p.networkID != nil {
		networkID = p.networkID.Uint64()
	}

	// Support networks like Goerli that use Proof of Authority as the default consensus mechanism.
	if value, exists := DefaultNetworkConsensusMechanism[networkID]; exists {
		consensusMechanism = value
	}

	if p.totalDifficulty.Cmp(&spec.TerminalTotalDifficulty) >= 0 {
		consensusMechanism = ProofOfStake
	}

	p.consensusMechanism.Reset()
	p.consensusMechanism.WithLabelValues(consensusMechanism.Name, consensusMechanism.Short).Set(consensusMechanism.Priority)

	return nil
}
