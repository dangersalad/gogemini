package gemini

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type GeminiAPI struct {
	BaseURL   string
	ApiKey    string
	ApiSecret string
	Nonce     int64
	logger    *log.Logger
}

type GeminiError struct {
	Result     string `json:"result"`
	Reason     string `json:"reason"`
	Message    string `json:"message"`
	StatusCode int
}

func (ge *GeminiError) String() string {
	return fmt.Sprintf("%s - %s", ge.Reason, ge.Message)
}

func (ge GeminiError) Error() string {
	return ge.String()
}

// Ticker stores the json returned by the pubticker endpoint
type Ticker struct {
	Bid  float64 `json:"bid,string"`
	Ask  float64 `json:"ask,string"`
	Last float64 `json:"last,string"`
}

// Fund stores the json returned by the funds endpoint
type Fund struct {
	Type                   string  `json:"type"`
	Currency               string  `json:"currency"`
	Amount                 float64 `json:"amount,string"`
	Available              float64 `json:"available,string"`
	AvailableForWithdrawal float64 `json:"availableForWithdrawal,string"`
}

// WithdrawResponse is the response from a fund withdraw request
type WithdrawResponse struct {
	Destination string  `json:"destination"`
	Amount      float64 `json:"amount,string"`
	TXID        string  `json:"txHash"`
}

func (w *WithdrawResponse) String() string {
	return fmt.Sprintf("Withdrew %0.8f to %s, tdix=%s", w.Amount, w.Destination, w.TXID)
}

// Order stores the json returned by placing an order or getting order status
type Order struct {
	OrderId         string  `json:"order_id"`
	ClientId        string  `json:"client_order_id"`
	Symbol          string  `json:"symbol"`
	Price           float64 `json:"price,string"`
	AvgExecPrice    float64 `json:"avg_execution_price,string"`
	Side            string  `json:"side"`
	Type            string  `json:"type"`
	Timestamp       int     `json:"timestamp,string"`
	TimestampMs     int     `json:"timestampms"`
	Live            bool    `json:"is_live"`
	Cancelled       bool    `json:"is_cancelled"`
	ExecutedAmount  float64 `json:"executed_amount,string"`
	RemainingAmount float64 `json:"remaining_amount,string"`
	OrigAmount      float64 `json:"original_amount,string"`
}

// OderbookOrder are the sub documents on the Orderbook response
type OrderbookOrder struct {
	Price     float64 `json:"price,string"`
	Amount    float64 `json:"amount,string"`
	Timestamp int     `json:"timestamp,string"`
}

// Orderbook stores the json returned by GetOrderbook
type Orderbook struct {
	Bids []OrderbookOrder `json:"bids"`
	Asks []OrderbookOrder `json:"asks"`
}

// Request is used to set the data for making an api request
type Request interface {
	SetNonce(int64)
	GetPayload() []byte
	GetRoute() string
}

type BaseRequest struct {
	Request string `json:"request"`
	Nonce   int64  `json:"nonce"`
}

func (r *BaseRequest) GetPayload() []byte {
	data, _ := json.Marshal(r)
	return data
}

func (r *BaseRequest) GetRoute() string {
	return r.Request
}

func (r *BaseRequest) SetNonce(n int64) {
	r.Nonce = n
}

func NewBaseRequest(route string) BaseRequest {
	return BaseRequest{
		Request: route,
	}
}

type OrderPlaceReq struct {
	BaseRequest
	Symbol   string   `json:"symbol"`
	Amount   string   `json:"amount"`
	Price    string   `json:"price"`
	Side     string   `json:"side"`
	Type     string   `json:"type"`
	ClientId string   `json:"client_order_id"`
	Options  []string `json:"options"`
}

func (r *OrderPlaceReq) GetPayload() []byte {
	data, _ := json.Marshal(r)
	return data
}

type WithdrawReq struct {
	BaseRequest
	Address string `json:"address"`
	Amount  string `json:"amount"`
}

func (r *WithdrawReq) GetPayload() []byte {
	data, _ := json.Marshal(r)
	return data
}

// AuthAPIReq makes a signed api request to gemini
func (ga *GeminiAPI) AuthAPIReq(r Request) ([]byte, error) {
	client := &http.Client{}
	r.SetNonce(ga.Nonce)
	ga.Nonce++
	reqURL := fmt.Sprintf("%s%s", ga.BaseURL, r.GetRoute())
	req, err := http.NewRequest("POST", reqURL, nil)
	if err != nil {
		ga.logger.Printf("ERROR: Failed to POST authenticated request to: %s\n", r.GetRoute())
		return []byte{}, nil
	}
	payload := r.GetPayload()
	ga.logger.Printf("Payload: %s\n", payload)
	base64Payload := base64.StdEncoding.EncodeToString(payload)
	h := hmac.New(sha512.New384, []byte(ga.ApiSecret))
	h.Write([]byte(base64Payload))
	sig := h.Sum(nil)
	req.Header.Add("X-GEMINI-APIKEY", ga.ApiKey)
	req.Header.Add("X-GEMINI-PAYLOAD", base64Payload)
	req.Header.Add("X-GEMINI-SIGNATURE", hex.EncodeToString(sig))
	resp, err := client.Do(req)
	if err != nil {
		ga.logger.Printf("ERROR: failed to POST authenticated request: %s\n", r.GetRoute())
		return []byte{}, nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ga.logger.Printf("ERROR: failed to read response body\n")
		return []byte{}, nil
	}

	// check for error
	if resp.StatusCode > 399 {
		geminiErr := &GeminiError{}
		err = json.Unmarshal(body, geminiErr)
		if err != nil {
			ga.logger.Printf("ERROR: error decoding json response\n")
			return nil, err
		}
		geminiErr.StatusCode = resp.StatusCode
		return nil, geminiErr
	}

	return body, nil
}

// GetTicker takes a ticker pair and returns a Ticker struct
func (ga *GeminiAPI) GetTicker(pair string) (Ticker, error) {
	tickerUrl := fmt.Sprintf("/v1/pubticker/%s", pair)
	resp, err := http.Get(fmt.Sprintf("%s%s", ga.BaseURL, tickerUrl))
	if err != nil {
		ga.logger.Printf("ERROR: Failed to get ticker for pair %s\n", pair)
		return Ticker{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ga.logger.Printf("ERROR: Failed to read ticker from response\n")
		return Ticker{}, err
	}
	ticker := Ticker{}
	err = json.Unmarshal(body, &ticker)
	if err != nil {
		ga.logger.Printf("ERROR: Failed to decode ticker from response\n")
		return ticker, err
	}
	return ticker, nil
}

// GetOrderbook takes a currency symbol and returns a slice of Order structs
func (ga *GeminiAPI) GetOrderbook(pair string, bidLimit, askLimit int) (Orderbook, error) {
	tickerUrl := fmt.Sprintf("/v1/book/%s?limit_bids=%d&limit_asks=%d", pair, bidLimit, askLimit)
	resp, err := http.Get(fmt.Sprintf("%s%s", ga.BaseURL, tickerUrl))
	if err != nil {
		ga.logger.Printf("ERROR: Failed to get ticker for pair %s\n", pair)
		return Orderbook{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ga.logger.Printf("ERROR: Failed to read ticker from response\n")
		return Orderbook{}, err
	}
	orders := Orderbook{}
	err = json.Unmarshal(body, &orders)
	if err != nil {
		ga.logger.Printf("ERROR: Failed to decode Orderbook from response: %s\n", body)
		return Orderbook{}, err
	}
	return orders, nil
}

// GetFunds returns a list of Fund structs
func (ga *GeminiAPI) GetFunds() ([]Fund, error) {
	input := NewBaseRequest("/v1/balances")
	body, err := ga.AuthAPIReq(&input)
	if err != nil {
		ga.logger.Printf("ERROR: Failed to get Funds\n")
		return []Fund{}, err
	}
	funds := []Fund{}
	err = json.Unmarshal(body, &funds)
	if err != nil {
		ga.logger.Printf("ERROR: Failed to get Funds\n")
		return []Fund{}, err
	}
	return funds, nil
}

// Withdraw send the specified amount of funds of the specified
// currency from your account to a specified address
func (ga *GeminiAPI) Withdraw(currency, address string, amount float64) (*WithdrawResponse, error) {
	amountStr := fmt.Sprintf("%0.8f", amount)
	input := &WithdrawReq{
		BaseRequest: NewBaseRequest(fmt.Sprintf("/v1/withdraw/%s", currency)),
		Address:     address,
		Amount:      amountStr,
	}
	body, err := ga.AuthAPIReq(input)
	if err != nil {
		ga.logger.Printf("ERROR: Failed to withdraw\n")
		return nil, err
	}
	resp := &WithdrawResponse{}
	err = json.Unmarshal(body, resp)
	if err != nil {
		ga.logger.Printf("ERROR: Failed to withdraw\n")
		return nil, err
	}
	return resp, nil
}

// GetOrderStatus returns a list of Order structs
func (ga *GeminiAPI) GetOrderStatus() ([]Order, error) {
	input := NewBaseRequest("/v1/orders")
	orders := []Order{}
	body, err := ga.AuthAPIReq(&input)
	if err != nil {
		ga.logger.Printf("ERROR: Failed to get order status\n")
		return []Order{}, err
	}
	err = json.Unmarshal(body, &orders)
	if err != nil {
		ga.logger.Printf("ERROR: Failed to decode order status json\n")
		return []Order{}, err
	}
	return orders, nil
}

// CancelAll attempts to cancel all open orders on the session
func (ga *GeminiAPI) CancelAll() {
	input := NewBaseRequest("/v1/order/cancel/session")
	ga.AuthAPIReq(&input)
}

// PlaceLimitOrder takes a direction, pair, client_id, amount, and price and returns an Order object
func (ga *GeminiAPI) PlaceLimitOrder(side, pair, client_id string, amount, price float64, options []string) (Order, error) {
	amountStr := fmt.Sprintf("%0.8f", amount)
	priceStr := ""
	if pair == "btcusd" || pair == "ethusd" {
		priceStr = fmt.Sprintf("%0.2f", price)
	} else if pair == "ethbtc" {
		priceStr = fmt.Sprintf("%0.5f", price)
	} else {
		return Order{}, errors.New("Unsupported pair for placing orders")
	}
	orderReq := &OrderPlaceReq{
		BaseRequest: NewBaseRequest("/v1/order/new"),
		Symbol:      pair,
		Amount:      amountStr,
		Price:       priceStr,
		Side:        side,
		Type:        "exchange limit",
		ClientId:    client_id,
		Options:     options,
	}

	body, err := ga.AuthAPIReq(orderReq)
	if err != nil {
		ga.logger.Printf("ERROR: error placing order\n")
		return Order{}, err
	}
	order := Order{}
	err = json.Unmarshal(body, &order)
	if err != nil {
		ga.logger.Printf("ERROR: error decoding order placement json response\n")
		return Order{}, err
	}
	return order, nil
}

// NewGeminiAPI initializes a GeminiAPI object
func NewGeminiAPI(baseurl, apikey, apisecret string, logger *log.Logger) *GeminiAPI {

	if logger == nil {
		logger = log.New(os.Stderr, "gemini api: ", log.Ldate|log.Ltime|log.Lshortfile)
	}

	ga := &GeminiAPI{
		BaseURL:   baseurl,
		ApiKey:    apikey,
		ApiSecret: apisecret,
		Nonce:     time.Now().UnixNano(),
		logger:    logger,
	}

	return ga
}
