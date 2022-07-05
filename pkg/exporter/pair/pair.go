package pair

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	eth2client "github.com/attestantio/go-eth2-client"
	"github.com/attestantio/go-eth2-client/http"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/onrik/ethrpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
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
	consensusClient eth2client.Service
	ethrpcClient    *ethrpc.EthRPC
	bootstrapped    bool
	consensusURL    string
	executionURL    string

	totalDifficulty         *big.Int
	terminalTotalDifficulty *big.Int
	networkID               *big.Int

	networkIDFetchedAt time.Time
	ttdFetchedAt       time.Time
	tdFetchedAt        time.Time
}

// NewMetrics returns a new Metrics instance.
func NewMetrics(ctx context.Context, log logrus.FieldLogger, namespace, consensusURL, executionURL string) (Metrics, error) {
	p := &pair{
		log: log,

		executionURL: executionURL,
		consensusURL: consensusURL,

		consensusClient: nil,
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

		totalDifficulty:         big.NewInt(0),
		terminalTotalDifficulty: big.NewInt(0),
	}

	prometheus.MustRegister(p.consensusMechanism)

	return p, nil
}

func (p *pair) Bootstrap(ctx context.Context) error {
	consenusClient, err := http.New(ctx,
		http.WithAddress(p.consensusURL),
		http.WithLogLevel(zerolog.Disabled),
	)
	if err != nil {
		return err
	}

	p.consensusClient = consenusClient

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

				if time.Since(p.ttdFetchedAt) > 15*time.Minute {
					if err := p.fetchTerminalTotalDifficulty(ctx); err != nil {
						p.log.WithError(err).Error("Failed to fetch terminal total difficulty")
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

func (p *pair) fetchTerminalTotalDifficulty(ctx context.Context) error {
	provider, isProvider := p.consensusClient.(eth2client.SpecProvider)
	if !isProvider {
		return errors.New("client does not implement eth2client.SpecProvider")
	}

	spec, err := provider.Spec(ctx)
	if err != nil {
		return err
	}

	terminalTotalDifficulty, exists := spec["TERMINAL_TOTAL_DIFFICULTY"]
	if !exists {
		return errors.New("TERMINAL_TOTAL_DIFFICULTY not found in spec")
	}

	ttd := cast.ToString(fmt.Sprintf("%v", terminalTotalDifficulty))

	asBigInt, success := big.NewInt(0).SetString(ttd, 10)
	if !success {
		return errors.New("TERMINAL_TOTAL_DIFFICULTY not a valid integer")
	}

	p.terminalTotalDifficulty = asBigInt

	p.ttdFetchedAt = time.Now()

	return nil
}

func (p *pair) deriveConsensusMechanism(ctx context.Context) error {
	if p.totalDifficulty == big.NewInt(0) {
		return errors.New("total difficulty not fetched")
	}

	if p.terminalTotalDifficulty == big.NewInt(0) {
		return errors.New("terminal total difficulty not fetched")
	}

	if p.networkID == big.NewInt(0) {
		return errors.New("network ID not fetched")
	}

	consensusMechanism := DefaultConsensusMechanism

	// Support networks like Goerli that use Proof of Authority as the default consensus mechanism.
	if value, exists := DefaultNetworkConsensusMechanism[p.networkID.Uint64()]; exists {
		consensusMechanism = value
	}

	if p.totalDifficulty.Cmp(p.terminalTotalDifficulty) >= 0 {
		consensusMechanism = ProofOfStake
	}

	p.consensusMechanism.Reset()
	p.consensusMechanism.WithLabelValues(consensusMechanism.Name, consensusMechanism.Short).Set(consensusMechanism.Priority)

	return nil
}
