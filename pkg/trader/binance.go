package trader

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	binance "github.com/adshao/go-binance/v2"
	"github.com/shopspring/decimal"
)

type BinanceTrader struct {
    Client *binance.Client
}

func NewBinanceTrader() *BinanceTrader {
    client := binance.NewClient(os.Getenv("BINANCE_API_KEY"), os.Getenv("BINANCE_SECRET_KEY"))
    return &BinanceTrader{Client: client}
}

func (t *BinanceTrader) GetBalances() map[string]string {
    account, err := t.Client.NewGetAccountService().Do(context.Background())
    if err != nil {
        log.Fatalf("Error getting account balances: %v", err)
        return nil
    }
    balances := map[string]string{}
    for _, balance := range account.Balances {
        assetBalance, _ := decimal.NewFromString(balance.Free)
        if balance.Free != "0" && assetBalance.Truncate(0).String() != "0" && !strings.HasPrefix(balance.Asset, "LD") {
            log.Printf("%s: %s, %s\n", balance.Asset, balance.Free, assetBalance.Truncate(0).String())
            balances[balance.Asset] = balance.Free
        }
    }
    return balances
}

// PlaceMarketBuyOrder places a market buy order using the amount in the counter currency
func (t *BinanceTrader) PlaceMarketBuyOrder(base, counter string, amount float64) {
	symbol := fmt.Sprintf("%s%s", base, counter)

	// Get the latest price for the symbol
	prices, err := t.Client.NewListPricesService().Symbol(symbol).Do(context.Background())
	if err != nil {
		log.Fatalf("Error getting price for %s: %v", symbol, err)
	}

	// Assuming only one price is returned
	price, err := strconv.ParseFloat(prices[0].Price, 64)
	if err != nil {
		log.Fatalf("Error parsing price: %v", err)
	}

	// Calculate the quantity in the base currency
	quantity := amount / price
  log.Printf("Price %s/%s: %f\n", base, counter, price)
  log.Printf("We are about to place %f %s for %f %s \n", quantity, base, amount, counter)

	order, err := t.Client.NewCreateOrderService().Symbol(symbol).
		Side(binance.SideTypeBuy).
		Type(binance.OrderTypeMarket).
		Quantity(fmt.Sprintf("%f", quantity)).
		Do(context.Background())

	if err != nil {
		log.Fatalf("Error placing market buy order: %v", err)
	}

	log.Printf("Market Buy Order placed: %+v\n", order)
}

// PlaceMarketSellOrder places a market sell order using the quantity in the base currency
func (t *BinanceTrader) PlaceMarketSellOrder(base, counter string, quantity float64) {
	symbol := fmt.Sprintf("%s%s", base, counter)

	order, err := t.Client.NewCreateOrderService().Symbol(symbol).
		Side(binance.SideTypeSell).
		Type(binance.OrderTypeMarket).
		Quantity(fmt.Sprintf("%f", quantity)).
		Do(context.Background())

	if err != nil {
		log.Fatalf("Error placing market sell order: %v", err)
	}

	log.Printf("Market Sell Order placed: %+v\n", order)
}
