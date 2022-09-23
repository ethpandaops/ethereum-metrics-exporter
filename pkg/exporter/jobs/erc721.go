package jobs

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/savid/ethereum-address-metrics-exporter/pkg/exporter/api"
	"github.com/sirupsen/logrus"
)

// ERC721 exposes metrics for ethereum ERC721 contract by address
type ERC721 struct {
	client        api.ExecutionClient
	log           logrus.FieldLogger
	ERC721Balance prometheus.GaugeVec
	addresses     []*AddressERC721
	labelsMap     map[string]int
}

type AddressERC721 struct {
	Address  string            `yaml:"address"`
	Contract string            `yaml:"contract"`
	Name     string            `yaml:"name"`
	Labels   map[string]string `yaml:"labels"`
}

const (
	NameERC721 = "erc721"
)

func (n *ERC721) Name() string {
	return NameERC721
}

// NewERC721 returns a new ERC721 instance.
func NewERC721(client api.ExecutionClient, log logrus.FieldLogger, namespace string, constLabels map[string]string, addresses []*AddressERC721) ERC721 {
	namespace += "_" + NameERC721

	labelsMap := map[string]int{
		LabelName:     0,
		LabelAddress:  1,
		LabelContract: 2,
	}

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

	instance := ERC721{
		client:    client,
		log:       log.WithField("module", NameERC721),
		addresses: addresses,
		labelsMap: labelsMap,
		ERC721Balance: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "balance",
				Help:        "The balance of a ethereum ERC721 contract by address.",
				ConstLabels: constLabels,
			},
			labels,
		),
	}

	prometheus.MustRegister(instance.ERC721Balance)

	return instance
}

func (n *ERC721) Start(ctx context.Context) {
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
func (n *ERC721) tick(ctx context.Context) {
	for _, address := range n.addresses {
		err := n.getBalance(address)

		if err != nil {
			n.log.WithError(err).WithField("address", address).Error("Failed to get erc721 contract balanceOf address")
		}
	}
}

func (n *ERC721) getLabelValues(address *AddressERC721) []string {
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
			case LabelContract:
				values[index] = address.Contract
			default:
				values[index] = LabelDefaultValue
			}
		}
	}

	return values
}

func (n *ERC721) getBalance(address *AddressERC721) error {
	// call balanceOf(address) which is 0x70a08231
	balanceOfData := "0x70a08231000000000000000000000000" + address.Address[2:]

	balanceStr, err := n.client.ETHCall(&api.ETHCallTransaction{
		To:   address.Contract,
		Data: &balanceOfData,
	}, "latest")
	if err != nil {
		return err
	}

	n.ERC721Balance.WithLabelValues(n.getLabelValues(address)...).Set(hexStringToFloat64(balanceStr))

	return nil
}
