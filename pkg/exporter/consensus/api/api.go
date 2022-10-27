package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ethpandaops/ethereum-metrics-exporter/pkg/exporter/consensus/api/types"
	"github.com/sirupsen/logrus"
)

// ConsensusClient is an interface for executing RPC calls to the Ethereum node.
type ConsensusClient interface {
	NodePeer(ctx context.Context, peerID string) (types.Peer, error)
	NodePeers(ctx context.Context) (types.Peers, error)
	NodePeerCount(ctx context.Context) (types.PeerCount, error)
}

type consensusClient struct {
	url    string
	log    logrus.FieldLogger
	client http.Client
}

// NewConsensusClient creates a new ConsensusClient.
func NewConsensusClient(ctx context.Context, log logrus.FieldLogger, url string) ConsensusClient {
	client := http.Client{
		Timeout: time.Second * 10,
	}

	return &consensusClient{
		url:    url,
		log:    log,
		client: client,
	}
}

type apiResponse struct {
	Data json.RawMessage `json:"data"`
}

//nolint:unparam // ctx will probably be used in the future
func (c *consensusClient) post(ctx context.Context, path string, body map[string]interface{}) (json.RawMessage, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	rsp, err := c.client.Post(c.url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", rsp.StatusCode)
	}

	data, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}

	resp := new(apiResponse)
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, err
	}

	return resp.Data, nil
}

//nolint:unparam // ctx will probably be used in the future
func (c *consensusClient) get(ctx context.Context, path string) (json.RawMessage, error) {
	rsp, err := c.client.Get(c.url + path)
	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", rsp.StatusCode)
	}

	data, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}

	resp := new(apiResponse)
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, err
	}

	return resp.Data, nil
}

func (c *consensusClient) NodePeers(ctx context.Context) (types.Peers, error) {
	data, err := c.get(ctx, "/eth/v1/node/peers")
	if err != nil {
		return nil, err
	}

	rsp := types.Peers{}
	if err := json.Unmarshal(data, &rsp); err != nil {
		return nil, err
	}

	return rsp, nil
}

func (c *consensusClient) NodePeer(ctx context.Context, peerID string) (types.Peer, error) {
	data, err := c.get(ctx, fmt.Sprintf("/eth/v1/node/peers/%s", peerID))
	if err != nil {
		return types.Peer{}, err
	}

	rsp := types.Peer{}
	if err := json.Unmarshal(data, &rsp); err != nil {
		return types.Peer{}, err
	}

	return rsp, nil
}

func (c *consensusClient) NodePeerCount(ctx context.Context) (types.PeerCount, error) {
	data, err := c.get(ctx, "/eth/v1/node/peer_count")
	if err != nil {
		return types.PeerCount{}, err
	}

	rsp := types.PeerCount{}
	if err := json.Unmarshal(data, &rsp); err != nil {
		return types.PeerCount{}, err
	}

	return rsp, nil
}
