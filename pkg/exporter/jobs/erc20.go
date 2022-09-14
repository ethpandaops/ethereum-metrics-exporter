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
}

type AddressERC20 struct {
	Address  string `yaml:"address"`
	Contract string `yaml:"contract"`
	Name     string `yaml:"name"`
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

	instance := ERC20{
		client:    client,
		log:       log.WithField("module", NameERC20),
		addresses: addresses,
		ERC20Balance: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "balance",
				Help:        "The balance of a ethereum ERC20 contract by address.",
				ConstLabels: constLabels,
			},
			[]string{"name", "address", "contract", "symbol"},
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

	n.ERC20Balance.WithLabelValues(address.Name, address.Address, address.Contract, symbol).Set(hexStringToFloat64(balanceStr))

	return nil
}
