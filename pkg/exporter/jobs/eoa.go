package jobs

import (
	"context"
	"math/big"
	"time"

	"github.com/onrik/ethrpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// Eth exposes metrics for ethereum externally owned account addresses
type EOA struct {
	client     *ethrpc.EthRPC
	log        logrus.FieldLogger
	EOABalance prometheus.GaugeVec
	addresses  []*AddressEOA
}

type AddressEOA struct {
	Address string `yaml:"address"`
	Name    string `yaml:"name"`
}

const (
	NameEOA = "eoa"
)

func (n *EOA) Name() string {
	return NameEOA
}

// NewEOA returns a new EOA instance.
func NewEOA(client *ethrpc.EthRPC, log logrus.FieldLogger, namespace string, constLabels map[string]string, addresses []*AddressEOA) EOA {
	namespace += "_" + NameEOA

	instance := EOA{
		client:    client,
		log:       log.WithField("module", NameEOA),
		addresses: addresses,
		EOABalance: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "balance",
				Help:        "The balance of a ethereum externally owned account address.",
				ConstLabels: constLabels,
			},
			[]string{"name", "address"},
		),
	}

	prometheus.MustRegister(instance.EOABalance)

	return instance
}

func (n *EOA) Start(ctx context.Context) {
	n.tick(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 15):
			n.tick(ctx)
		}
	}
}

//nolint:unparam // context will be used in the future
func (n *EOA) tick(ctx context.Context) {
	for _, address := range n.addresses {
		err := n.getBalance(address)

		if err != nil {
			n.log.WithError(err).WithField("address", address).Error("Failed to get EOA balance")
		}
	}
}

func (n *EOA) getBalance(address *AddressEOA) error {
	balance, err := n.client.EthGetBalance(address.Address, "latest")
	if err != nil {
		return err
	}

	balanceFloat64, _ := new(big.Float).SetInt(&balance).Float64()
	n.EOABalance.WithLabelValues(address.Name, address.Address).Set(balanceFloat64)

	return nil
}
