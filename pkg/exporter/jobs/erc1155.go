package jobs

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethpandaops/ethereum-address-metrics-exporter/pkg/exporter/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// ERC1155 exposes metrics for ethereum ERC115 contract by address and token id
type ERC1155 struct {
	client         api.ExecutionClient
	log            logrus.FieldLogger
	ERC1155Balance prometheus.GaugeVec
	addresses      []*AddressERC1155
	labelsMap      map[string]int
}

type AddressERC1155 struct {
	Address  string            `yaml:"address"`
	Contract string            `yaml:"contract"`
	TokenID  big.Int           `yaml:"tokenID"`
	Name     string            `yaml:"name"`
	Labels   map[string]string `yaml:"labels"`
}

const (
	NameERC1155 = "erc1155"
)

func (n *ERC1155) Name() string {
	return NameERC1155
}

// NewERC1155 returns a new ERC1155 instance.
func NewERC1155(client api.ExecutionClient, log logrus.FieldLogger, namespace string, constLabels map[string]string, addresses []*AddressERC1155) ERC1155 {
	namespace += "_" + NameERC1155

	labelsMap := map[string]int{
		LabelName:     0,
		LabelAddress:  1,
		LabelContract: 2,
		LabelTokenID:  3,
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

	instance := ERC1155{
		client:    client,
		log:       log.WithField("module", NameERC1155),
		addresses: addresses,
		labelsMap: labelsMap,
		ERC1155Balance: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "balance",
				Help:        "The balance of a ethereum ERC115 contract by address and token id.",
				ConstLabels: constLabels,
			},
			labels,
		),
	}

	prometheus.MustRegister(instance.ERC1155Balance)

	return instance
}

func (n *ERC1155) Start(ctx context.Context) {
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
func (n *ERC1155) tick(ctx context.Context) {
	for _, address := range n.addresses {
		err := n.getBalance(address)

		if err != nil {
			n.log.WithError(err).WithField("address", address).Error("Failed to get erc1155 contract balanceOf address")
		}
	}
}

func (n *ERC1155) getLabelValues(address *AddressERC1155) []string {
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
			case LabelTokenID:
				values[index] = address.TokenID.String()
			default:
				values[index] = LabelDefaultValue
			}
		}
	}

	return values
}

func (n *ERC1155) getBalance(address *AddressERC1155) error {
	// call balanceOf(address,uint256) which is 0x00fdd58e
	balanceOfData := "0x00fdd58e000000000000000000000000" + address.Address[2:] + fmt.Sprintf("%064x", &address.TokenID)

	balanceStr, err := n.client.ETHCall(&api.ETHCallTransaction{
		To:   address.Contract,
		Data: &balanceOfData,
	}, "latest")
	if err != nil {
		return err
	}

	n.ERC1155Balance.WithLabelValues(n.getLabelValues(address)...).Set(hexStringToFloat64(balanceStr))

	return nil
}
