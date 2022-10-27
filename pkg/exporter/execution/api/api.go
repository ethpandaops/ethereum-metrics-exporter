package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethpandaops/ethereum-metrics-exporter/pkg/exporter/execution/api/types"
	"github.com/sirupsen/logrus"
)

// ExecutionClient is an interface for executing RPC calls to the Ethereum node.
type ExecutionClient interface {
	// AdminNodeInfo returns information about the node.
	AdminNodeInfo(ctx context.Context) (*types.NodeInfo, error)
	// AdminPeers returns information about the peers.
	AdminPeers(ctx context.Context) ([]*p2p.PeerInfo, error)
	// TXPoolStatus returns information about the transaction pool.
	TXPoolStatus(ctx context.Context) (*types.TXPoolStatus, error)
	// NetPeerCount returns the number of peers.
	NetPeerCount(ctx context.Context) (int, error)
}

type executionClient struct {
	url    string
	log    logrus.FieldLogger
	client http.Client
}

// NewExecutionClient creates a new ExecutionClient.
func NewExecutionClient(ctx context.Context, log logrus.FieldLogger, url string) ExecutionClient {
	client := http.Client{
		Timeout: time.Second * 10,
	}

	return &executionClient{
		url:    url,
		log:    log,
		client: client,
	}
}

type apiResponse struct {
	JSONRpc string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result"`
}

//nolint:unparam // ctx will probably be used in the future
func (e *executionClient) post(ctx context.Context, method string, params []string, id int) (json.RawMessage, error) {
	body := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"id":      id,
		"params":  params,
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	rsp, err := e.client.Post(e.url, "application/json", bytes.NewBuffer(jsonData))
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

	return resp.Result, nil
}

func (e *executionClient) AdminNodeInfo(ctx context.Context) (*types.NodeInfo, error) {
	rsp, err := e.post(ctx, "admin_nodeInfo", []string{}, 0)
	if err != nil {
		return nil, err
	}

	nodeInfo := &types.NodeInfo{}
	if err := json.Unmarshal(rsp, nodeInfo); err != nil {
		return nil, err
	}

	return nodeInfo, nil
}

func (e *executionClient) AdminPeers(ctx context.Context) ([]*p2p.PeerInfo, error) {
	rsp, err := e.post(ctx, "admin_peers", []string{}, 0)
	if err != nil {
		return nil, err
	}

	peers := []*p2p.PeerInfo{}
	if err := json.Unmarshal(rsp, &peers); err != nil {
		return nil, err
	}

	return peers, nil
}

func (e *executionClient) NetPeerCount(ctx context.Context) (int, error) {
	rsp, err := e.post(ctx, "net_peerCount", []string{}, 0)
	if err != nil {
		return 0, err
	}

	count := hexutil.Uint64(0)
	if err := json.Unmarshal(rsp, &count); err != nil {
		return 0, err
	}

	return int(count), nil
}

func (e *executionClient) TXPoolStatus(ctx context.Context) (*types.TXPoolStatus, error) {
	rsp, err := e.post(ctx, "txpool_status", []string{}, 0)
	if err != nil {
		return nil, err
	}

	txPoolStatus := &types.TXPoolStatus{}
	if err := json.Unmarshal(rsp, txPoolStatus); err != nil {
		return nil, err
	}

	return txPoolStatus, nil
}
