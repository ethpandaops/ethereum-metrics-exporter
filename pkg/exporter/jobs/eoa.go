package jobs

import (
	"context"
	"time"

	"github.com/ethpandaops/ethereum-address-metrics-exporter/pkg/exporter/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// Eth exposes metrics for ethereum externally owned account addresses
type EOA struct {
	client     api.ExecutionClient
	log        logrus.FieldLogger
	EOABalance prometheus.GaugeVec
	addresses  []*AddressEOA
	labelsMap  map[string]int
}

type AddressEOA struct {
	Address string            `yaml:"address"`
	Name    string            `yaml:"name"`
	Labels  map[string]string `yaml:"labels"`
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

	labelsMap := map[string]int{}
	labelsMap[LabelName] = 0
	labelsMap[LabelAddress] = 1

	for address := range addresses {
		for label := range addresses[address].Labels {
			if _, ok := labelsMap[label]; !ok {
				labelsMap[label] = len(labelsMap)
			}
		}
	}

	labels := make([]string, len(labelsMap))
	for label, index := range labelsMap {
		labels[index] = label
	}

	instance := EOA{
		client:    client,
		log:       log.WithField("module", NameEOA),
		addresses: addresses,
		labelsMap: labelsMap,
		EOABalance: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "balance",
				Help:        "The balance of a ethereum externally owned account address.",
				ConstLabels: constLabels,
			},
			labels,
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

func (n *EOA) getLabelValues(address *AddressEOA) []string {
	values := make([]string, len(n.labelsMap))

	for label, index := range n.labelsMap {
		if address.Labels != nil && address.Labels[label] != "" {
			values[index] = address.Labels[label]
		} else {
			switch label {
			case LabelName:
				values[index] = address.Name
			case LabelAddress:
				values[index] = address.Address
			default:
				values[index] = LabelDefaultValue
			}
		}
	}

	return values
}

func (n *EOA) getBalance(address *AddressEOA) error {
	balance, err := n.client.ETHGetBalance(address.Address, "latest")
	if err != nil {
		return err
	}

	balanceFloat64 := hexStringToFloat64(balance)
	n.EOABalance.WithLabelValues(n.getLabelValues(address)...).Set(balanceFloat64)

	return nil
}
