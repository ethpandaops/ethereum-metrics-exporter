package jobs

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/onrik/ethrpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// ERC1155 exposes metrics for ethereum ERC115 contract by address and token id
type ERC1155 struct {
	client         *ethrpc.EthRPC
	log            logrus.FieldLogger
	ERC1155Balance prometheus.GaugeVec
	addresses      []*AddressERC1155
}

type AddressERC1155 struct {
	Address  string  `yaml:"address"`
	Contract string  `yaml:"contract"`
	TokenID  big.Int `yaml:"tokenID"`
	Name     string  `yaml:"name"`
}

const (
	NameERC1155 = "erc1155"
)

func (n *ERC1155) Name() string {
	return NameERC1155
}

// NewERC1155 returns a new ERC1155 instance.
func NewERC1155(client *ethrpc.EthRPC, log logrus.FieldLogger, namespace string, constLabels map[string]string, addresses []*AddressERC1155) ERC1155 {
	namespace += "_" + NameERC1155

	instance := ERC1155{
		client:    client,
		log:       log.WithField("module", NameERC1155),
		addresses: addresses,
		ERC1155Balance: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "balance",
				Help:        "The balance of a ethereum ERC115 contract by address and token id.",
				ConstLabels: constLabels,
			},
			[]string{"name", "address", "contract", "token_id"},
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

func (n *ERC1155) getBalance(address *AddressERC1155) error {
	balanceStr, err := n.client.EthCall(ethrpc.T{
		To:   address.Contract,
		From: "0x0000000000000000000000000000000000000000",
		// call balanceOf(address,uint256) which is 0x00fdd58e
		Data: "0x00fdd58e000000000000000000000000" + address.Address[2:] + fmt.Sprintf("%064x", &address.TokenID),
	}, "latest")
	if err != nil {
		return err
	}

	n.ERC1155Balance.WithLabelValues(address.Name, address.Address, address.Contract, address.TokenID.String()).Set(hexStringToFloat64(balanceStr))

	return nil
}
