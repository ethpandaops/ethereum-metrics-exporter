package jobs

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/savid/ethereum-address-metrics-exporter/pkg/exporter/api"
	"github.com/sirupsen/logrus"
)

// Eth exposes metrics for ethereum externally owned account addresses
type EOA struct {
	client     api.ExecutionClient
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
func NewEOA(client api.ExecutionClient, log logrus.FieldLogger, namespace string, constLabels map[string]string, addresses []*AddressEOA) EOA {
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
	balance, err := n.client.ETHGetBalance(address.Address, "latest")
	if err != nil {
		return err
	}

	balanceFloat64 := hexStringToFloat64(balance)
	n.EOABalance.WithLabelValues(address.Name, address.Address).Set(balanceFloat64)

	return nil
}
