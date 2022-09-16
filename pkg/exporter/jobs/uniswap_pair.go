package jobs

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/savid/ethereum-address-metrics-exporter/pkg/exporter/api"
	"github.com/sirupsen/logrus"
)

// UniswapPair exposes metrics for ethereum uniswap pair contract
type UniswapPair struct {
	client             api.ExecutionClient
	log                logrus.FieldLogger
	UniswapPairBalance prometheus.GaugeVec
	addresses          []*AddressUniswapPair
}

type AddressUniswapPair struct {
	From     string `yaml:"from"`
	To       string `yaml:"to"`
	Contract string `yaml:"contract"`
	Name     string `yaml:"name"`
}

const (
	NameUniswapPair = "uniswap_pair"
)

func (n *UniswapPair) Name() string {
	return NameUniswapPair
}

// NewUniswapPair returns a new UniswapPair instance.
func NewUniswapPair(client api.ExecutionClient, log logrus.FieldLogger, namespace string, constLabels map[string]string, addresses []*AddressUniswapPair) UniswapPair {
	namespace += "_" + NameUniswapPair

	instance := UniswapPair{
		client:    client,
		log:       log.WithField("module", NameUniswapPair),
		addresses: addresses,
		UniswapPairBalance: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "balance",
				Help:        "The balance of a ethereum uniswap pair contract.",
				ConstLabels: constLabels,
			},
			[]string{"name", "contract", "from", "to"},
		),
	}

	prometheus.MustRegister(instance.UniswapPairBalance)

	return instance
}

func (n *UniswapPair) Start(ctx context.Context) {
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
func (n *UniswapPair) tick(ctx context.Context) {
	for _, address := range n.addresses {
		err := n.getBalance(address)

		if err != nil {
			n.log.WithError(err).WithField("address", address).Error("Failed to get uniswap pair balance")
		}
	}
}

func (n *UniswapPair) getBalance(address *AddressUniswapPair) error {
	// call getReserves() which is 0x0902f1ac
	getReservesData := "0x0902f1ac000000000000000000000000"

	balanceStr, err := n.client.ETHCall(&api.ETHCallTransaction{
		To:   address.Contract,
		Data: &getReservesData,
	}, "latest")
	if err != nil {
		return err
	}

	if len(balanceStr) < 130 {
		n.log.WithFields(logrus.Fields{
			"address": address,
			"balance": balanceStr,
		}).Warn("Got empty uniswap pair balance")

		return nil
	}

	fromBalance := hexStringToFloat64(balanceStr[0:66])
	toBalance := hexStringToFloat64("0x" + balanceStr[66:130])

	balance := toBalance / fromBalance
	n.UniswapPairBalance.WithLabelValues(address.Name, address.Contract, address.From, address.To).Set(balance)

	return nil
}
