package jobs

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/savid/ethereum-address-metrics-exporter/pkg/exporter/api"
	"github.com/sirupsen/logrus"
)

// ERC20 exposes metrics for ethereum ERC20 contract by address
type ERC20 struct {
	client       api.ExecutionClient
	log          logrus.FieldLogger
	ERC20Balance prometheus.GaugeVec
	addresses    []*AddressERC20
	labelsMap    map[string]int
}

type AddressERC20 struct {
	Address  string            `yaml:"address"`
	Contract string            `yaml:"contract"`
	Name     string            `yaml:"name"`
	Labels   map[string]string `yaml:"labels"`
}

const (
	NameERC20 = "erc20"
)

func (n *ERC20) Name() string {
	return NameERC20
}

// NewERC20 returns a new ERC20 instance.
func NewERC20(client api.ExecutionClient, log logrus.FieldLogger, namespace string, constLabels map[string]string, addresses []*AddressERC20) ERC20 {
	namespace += "_" + NameERC20

	labelsMap := map[string]int{
		LabelName:     0,
		LabelAddress:  1,
		LabelContract: 2,
		LabelSymbol:   3,
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

	instance := ERC20{
		client:    client,
		log:       log.WithField("module", NameERC20),
		addresses: addresses,
		labelsMap: labelsMap,
		ERC20Balance: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "balance",
				Help:        "The balance of a ethereum ERC20 contract by address.",
				ConstLabels: constLabels,
			},
			labels,
		),
	}

	prometheus.MustRegister(instance.ERC20Balance)

	return instance
}

func (n *ERC20) Start(ctx context.Context) {
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
func (n *ERC20) tick(ctx context.Context) {
	for _, address := range n.addresses {
		err := n.getBalance(address)

		if err != nil {
			n.log.WithError(err).WithField("address", address).Error("Failed to get erc20 contract balanceOf address")
		}
	}
}

func (n *ERC20) getLabelValues(address *AddressERC20, symbol string) []string {
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
			case LabelSymbol:
				values[index] = symbol
			default:
				values[index] = LabelDefaultValue
			}
		}
	}

	return values
}

func (n *ERC20) getBalance(address *AddressERC20) error {
	// call balanceOf(address) which is 0x70a08231
	balanceOfData := "0x70a08231000000000000000000000000" + address.Address[2:]

	balanceStr, err := n.client.ETHCall(&api.ETHCallTransaction{
		To:   address.Contract,
		Data: &balanceOfData,
	}, "latest")
	if err != nil {
		return err
	}

	// call symbol() which is 0x95d89b41
	symbolData := "0x95d89b41000000000000000000000000"

	symbolHex, err := n.client.ETHCall(&api.ETHCallTransaction{
		To:   address.Contract,
		Data: &symbolData,
	}, "latest")
	if err != nil {
		return err
	}

	symbol, err := hexStringToString(symbolHex)
	if err != nil {
		return err
	}

	n.ERC20Balance.WithLabelValues(n.getLabelValues(address, symbol)...).Set(hexStringToFloat64(balanceStr))

	return nil
}
