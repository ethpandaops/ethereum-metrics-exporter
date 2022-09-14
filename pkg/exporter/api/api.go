package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// ExecutionClient is an interface for executing RPC calls to the Ethereum node.
type ExecutionClient interface {
	// ETHCall executes a new message call immediately without creating a transaction on the block chain.
	ETHCall(transaction *ETHCallTransaction, block string) (string, error)
	// ETHGetBalance returns the balance of the account of given address.
	ETHGetBalance(address string, block string) (string, error)
}

type ETHCallTransaction struct {
	From     *string `json:"from"`
	To       string  `json:"to"`
	Gas      *string `json:"gas"`
	GasPrice *string `json:"gasPrice"`
	Value    *string `json:"value"`
	Data     *string `json:"data"`
}

type executionClient struct {
	url     string
	log     logrus.FieldLogger
	client  http.Client
	headers map[string]string
}

// NewExecutionClient creates a new ExecutionClient.
func NewExecutionClient(log logrus.FieldLogger, url string, headers map[string]string, timeout time.Duration) ExecutionClient {
	client := http.Client{
		Timeout: timeout,
	}

	return &executionClient{
		url:     url,
		log:     log,
		client:  client,
		headers: headers,
	}
}

type apiResponse struct {
	JSONRpc string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result"`
}

//nolint:unparam // ctx will probably be used in the future
func (e *executionClient) post(method string, params interface{}, id int) (json.RawMessage, error) {
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

	req, err := http.NewRequest("POST", e.url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	for k, v := range e.headers {
		req.Header.Set(k, v)
	}

	req.Header.Set("Content-Type", "application/json")

	rsp, err := e.client.Do(req)
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

func (e *executionClient) ETHCall(transaction *ETHCallTransaction, block string) (string, error) {
	params := []interface{}{
		transaction,
		block,
	}

	rsp, err := e.post("eth_call", params, 1)
	if err != nil {
		return "", err
	}

	ethCall := ""
	if err := json.Unmarshal(rsp, &ethCall); err != nil {
		return "", err
	}

	return ethCall, nil
}

func (e *executionClient) ETHGetBalance(address, block string) (string, error) {
	params := []interface{}{
		address,
		block,
	}

	rsp, err := e.post("eth_getBalance", params, 1)
	if err != nil {
		return "", err
	}

	ethGetBalance := ""
	if err := json.Unmarshal(rsp, &ethGetBalance); err != nil {
		return "", err
	}

	return ethGetBalance, nil
}
