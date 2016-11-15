package gemini

import (
	"fmt"
	"testing"
)

const (
	url       = "https://api.sandbox.gemini.com/"
	apikey    = "<api key goes here>"
	apisecret = "<api secret goes here>"
)

func TestTicker(t *testing.T) {
	ga := NewGeminiAPI(url, "", "", nil)
	ticker, err := ga.GetTicker("BTCUSD")
	if err != nil {
		t.Fail()
	}
	fmt.Println(ticker)
}

func TestFunds(t *testing.T) {
	ga := NewGeminiAPI(url, apikey, apisecret, nil)
	funds, err := ga.GetFunds()
	if err != nil {
		t.Fail()
	}
	fmt.Println(funds)
}

func TestOrderStatus(t *testing.T) {
	ga := NewGeminiAPI(url, apikey, apisecret, nil)
	orders, err := ga.GetOrderStatus()
	if err != nil {
		t.Fail()
	}
	fmt.Println(orders)
}

func TestPlaceLimitOrder(t *testing.T) {
	ga := NewGeminiAPI(url, apikey, apisecret, nil)
	order, err := ga.PlaceLimitOrder("buy", "btcusd", "order1", 1.0, 1.0)
	if err != nil {
		t.Fail()
	}
	fmt.Println(order)
	order, err = ga.PlaceLimitOrder("sell", "btcusd", "order1", 1.0, 1.0)
	if err != nil {
		t.Fail()
	}
	fmt.Println(order)
}

func TestCancelAll(t *testing.T) {
	ga := NewGeminiAPI(url, apikey, apisecret, nil)
	ga.CancelAll()
}
