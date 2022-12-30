package price

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

var (
	binanceExchangeInfoRESTURL = "https://eapi.binance.com/eapi/v1/exchangeInfo"
	binanceOrderBookRESTURL    = "https://eapi.binance.com/eapi/v1/depth"
	binanceIndexPriceRESTURL   = "https://eapi.binance.com/eapi/v1/index"
)

var BaseCurrencies = []string{"BTC", "ETH"}
var BaseQuoteCurrency = "USDT"

type Index struct {
	Price float64 `json:"indexPrice,string"`
}

func GetCurrentSpotPrice(symbol string) (*Index, error) {
	resp, err := http.Get(fmt.Sprintf("%s?underlying=%s", binanceIndexPriceRESTURL, symbol))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var price Index
	if err := json.Unmarshal(body, &price); err != nil {
		return nil, err
	}

	return &price, nil
}

func GetPricesByCurrency() map[string]float64 {
	result := make(map[string]float64)
	for _, currency := range BaseCurrencies {
		index, err := GetCurrentSpotPrice(currency+BaseQuoteCurrency)
		if err != nil {
			log.Fatalf("Error %v\n", err)
		}
		result[currency] = index.Price
	}
	return result
}

func PricesMessage() string {
	result := ""
	pricesMap := GetPricesByCurrency()
	for currency, price := range pricesMap {
		result += fmt.Sprintf("%s price at the moment: %f %s\n", currency, price, BaseQuoteCurrency)
	}
	return result
}
