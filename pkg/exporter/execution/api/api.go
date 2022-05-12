package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/samcm/ethereum-metrics-exporter/pkg/exporter/execution/api/types"
	"github.com/sirupsen/logrus"
)

type ExecutionClient interface {
	AdminNodeInfo(ctx context.Context) (*types.NodeInfo, error)
	AdminPeers(ctx context.Context) ([]*p2p.PeerInfo, error)
	TXPoolStatus(ctx context.Context) (*types.TXPoolStatus, error)
	NetPeerCount(ctx context.Context) (int, error)
}

type executionClient struct {
	url    string
	log    logrus.FieldLogger
	client http.Client
}

func NewExecutionClient(ctx context.Context, log logrus.FieldLogger, url string) *executionClient {
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

func (e *executionClient) post(ctx context.Context, method string, params []string, id int) (json.RawMessage, error) {
	body := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"id":      id,
		"params":  params,
	}

	json_data, err := json.Marshal(body)
	if err != nil {
		log.Fatal(err)
	}

	rsp, err := e.client.Post(e.url, "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", rsp.StatusCode)
	}

	data, err := ioutil.ReadAll(rsp.Body)
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
