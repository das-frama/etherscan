package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	BaseURL *url.URL
	APIKey  string

	httpClient *http.Client
}

type ResponseTranscations struct {
	Status  string        `json:"status"`
	Message string        `json:"message"`
	Result  []Transcation `json:"result"`
}
type ResponseBalance struct {
	Status  string    `json:"status"`
	Message string    `json:"message"`
	Result  []Balance `json:"result"`
}

type Balance struct {
	Account string `json:"account"`
	Balance string `json:"balance"`
}

type Transcation struct {
	TimeStamp string `json:"timeStamp"`
	From      string `json:"from"`
}

// NormalTranscations get a list of "ERC20 - Token Transfer Events" by Address
// or get transfer events for a specific token contract
func (c *Client) NormalTranscations(address string, sortAsc bool, page, offset int) ([]Transcation, error) {
	sort := "desc"
	if sortAsc {
		sort = "asc"
	}
	req, err := c.newRequest(map[string]string{
		"module":  "account",
		"action":  "txlist",
		"address": address,
		"sort":    sort,
		"apiKey":  c.APIKey,
		"page":    strconv.Itoa(page),
		"offset":  strconv.Itoa(offset),
	})

	response := &ResponseTranscations{}
	_, err = c.do(req, &response)

	if response.Status != "1" && response.Message != "No transactions found" {
		err = fmt.Errorf("%s", response.Message)
		return nil, err
	}

	return response.Result, err
}

func (c *Client) BalanceMulti(addresses []string) ([]Balance, error) {
	address := strings.Join(addresses, ",")
	req, err := c.newRequest(map[string]string{
		"module":  "account",
		"action":  "balancemulti",
		"address": address,
		"apiKey":  c.APIKey,
	})

	response := &ResponseBalance{}
	_, err = c.do(req, &response)

	if response.Status != "1" && response.Message != "No transactions found" {
		err = fmt.Errorf("%s", response.Message)
		return nil, err
	}

	return response.Result, err
}

func (c *Client) newRequest(params map[string]string) (*http.Request, error) {
	q := c.BaseURL.Query()
	for key, val := range params {
		if val != "" {
			q.Set(key, val)
		}
	}
	c.BaseURL.RawQuery = q.Encode()
	// fmt.Println(c.BaseURL.RawQuery)
	req, err := http.NewRequest("GET", c.BaseURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(v)

	return resp, err
}

func NewClient(apiKey string) *Client {
	baseURL, err := url.Parse("http://api.etherscan.io/api")
	if err != nil {
		log.Fatal(err)
	}
	httpClient := &http.Client{
		Timeout: time.Duration(time.Second * 50),
	}

	return &Client{
		APIKey:     apiKey,
		BaseURL:    baseURL,
		httpClient: httpClient,
	}
}
