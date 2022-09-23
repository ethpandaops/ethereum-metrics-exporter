package jobs

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/savid/ethereum-address-metrics-exporter/pkg/exporter/api"
	"github.com/sirupsen/logrus"
)

// ChainlinkDataFeed exposes metrics for ethereum chainlink data feed contract
type ChainlinkDataFeed struct {
	client                   api.ExecutionClient
	log                      logrus.FieldLogger
	ChainlinkDataFeedBalance prometheus.GaugeVec
	addresses                []*AddressChainlinkDataFeed
	labelsMap                map[string]int
}

type AddressChainlinkDataFeed struct {
	From     string            `yaml:"from"`
	To       string            `yaml:"to"`
	Contract string            `yaml:"contract"`
	Name     string            `yaml:"name"`
	Labels   map[string]string `yaml:"labels"`
}

const (
	NameChainlinkDataFeed = "chainlink_data_feed"
)

func (n *ChainlinkDataFeed) Name() string {
	return NameChainlinkDataFeed
}

// NewChainlinkDataFeed returns a new ChainlinkDataFeed instance.
func NewChainlinkDataFeed(client api.ExecutionClient, log logrus.FieldLogger, namespace string, constLabels map[string]string, addresses []*AddressChainlinkDataFeed) ChainlinkDataFeed {
	namespace += "_" + NameChainlinkDataFeed

	labelsMap := map[string]int{
		LabelName:     0,
		LabelContract: 1,
		LabelFrom:     2,
		LabelTo:       3,
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

	instance := ChainlinkDataFeed{
		client:    client,
		log:       log.WithField("module", NameChainlinkDataFeed),
		addresses: addresses,
		labelsMap: labelsMap,
		ChainlinkDataFeedBalance: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "balance",
				Help:        "The balance of a ethereum chainlink data feed contract.",
				ConstLabels: constLabels,
			},
			labels,
		),
	}

	prometheus.MustRegister(instance.ChainlinkDataFeedBalance)

	return instance
}

func (n *ChainlinkDataFeed) Start(ctx context.Context) {
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
func (n *ChainlinkDataFeed) tick(ctx context.Context) {
	for _, address := range n.addresses {
		err := n.getBalance(address)

		if err != nil {
			n.log.WithError(err).WithField("address", address).Error("Failed to get chainlink data feed balance")
		}
	}
}

func (n *ChainlinkDataFeed) getLabelValues(address *AddressChainlinkDataFeed) []string {
	values := make([]string, len(n.labelsMap))

	for label, index := range n.labelsMap {
		if address.Labels != nil && address.Labels[label] != "" {
			values[index] = address.Labels[label]
		} else {
			switch label {
			case LabelName:
				values[index] = address.Name
			case LabelContract:
				values[index] = address.Contract
			case LabelFrom:
				values[index] = address.From
			case LabelTo:
				values[index] = address.To
			default:
				values[index] = LabelDefaultValue
			}
		}
	}

	return values
}

func (n *ChainlinkDataFeed) getBalance(address *AddressChainlinkDataFeed) error {
	// call latestAnswer() which is 0x50d25bcd
	latestAnswerData := "0x50d25bcd000000000000000000000000"

	balanceStr, err := n.client.ETHCall(&api.ETHCallTransaction{
		To:   address.Contract,
		Data: &latestAnswerData,
	}, "latest")
	if err != nil {
		return err
	}

	n.ChainlinkDataFeedBalance.WithLabelValues(n.getLabelValues(address)...).Set(hexStringToFloat64(balanceStr))

	return nil
}
