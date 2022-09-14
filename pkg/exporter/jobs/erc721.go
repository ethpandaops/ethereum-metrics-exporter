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
}

type AddressERC721 struct {
	Address  string `yaml:"address"`
	Contract string `yaml:"contract"`
	Name     string `yaml:"name"`
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

	instance := ERC721{
		client:    client,
		log:       log.WithField("module", NameERC721),
		addresses: addresses,
		ERC721Balance: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "balance",
				Help:        "The balance of a ethereum ERC721 contract by address.",
				ConstLabels: constLabels,
			},
			[]string{"name", "address", "contract"},
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

	n.ERC721Balance.WithLabelValues(address.Name, address.Address, address.Contract).Set(hexStringToFloat64(balanceStr))

	return nil
}
