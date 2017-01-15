package gemini

import (
	"fmt"
	"testing"
)

const (
	url       = "https://api.sandbox.gemini.com/"
	apikey    = "<api key>"
	apisecret = "<api secret>"
)

func TestTicker(t *testing.T) {
	ga := NewGeminiAPI(url, "", "", nil)
	_, err := ga.GetTicker("btcusd")
	if err != nil {
		t.Fail()
	}
}

func TestOrderbook(t *testing.T) {
	ga := NewGeminiAPI(url, "", "", nil)
	_, err := ga.GetOrderbook("btcusd", 1, 1)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
}

func TestFunds(t *testing.T) {
	ga := NewGeminiAPI(url, apikey, apisecret, nil)
	_, err := ga.GetFunds()
	if err != nil {
		t.Fail()
	}
}

func TestOrderStatus(t *testing.T) {
	ga := NewGeminiAPI(url, apikey, apisecret, nil)
	_, err := ga.GetOrderStatus()
	if err != nil {
		t.Fail()
	}
}

func TestPlaceLimitOrder(t *testing.T) {
	ga := NewGeminiAPI(url, apikey, apisecret, nil)
	_, err := ga.PlaceLimitOrder("buy", "btcusd", "order1", 1.0, 1.0, []string{"immediate-or-cancel"})
	if err != nil {
		t.Fail()
	}
	_, err = ga.PlaceLimitOrder("sell", "btcusd", "order1", 1.0, 1.0, []string{"immediate-or-cancel"})
	if err != nil {
		t.Fail()
	}
}

func TestWithdraw(t *testing.T) {
	ga := NewGeminiAPI(url, apikey, apisecret, nil)
	_, err := ga.Withdraw("btc", "1DFCqM24Sg4mKJqXPDLmPsF2hCGZkXwVff", 0.1)
	if err != nil {
		t.Fail()
	}
}

func TestBalances(t *testing.T) {
	ga := NewGeminiAPI(url, apikey, apisecret, nil)
	_, err := ga.GetBalance()
	if err != nil {
		t.Fail()
	}
}

func TestCancelAll(t *testing.T) {
	ga := NewGeminiAPI(url, apikey, apisecret, nil)
	ga.CancelAll()
}
