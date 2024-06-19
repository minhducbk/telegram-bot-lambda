package trader

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	binance "github.com/adshao/go-binance/v2"
	"github.com/shopspring/decimal"
)

const maxRetries = 10

type BinanceTrader struct {
    Client *binance.Client
}

func NewBinanceTrader() *BinanceTrader {
    client := binance.NewClient(os.Getenv("BINANCE_API_KEY"), os.Getenv("BINANCE_SECRET_KEY"))
    return &BinanceTrader{Client: client}
}

func (t *BinanceTrader) GetBalances() map[string]string {
	const retryDelay = 2 * time.Second

	var account *binance.Account
	var prices []*binance.SymbolPrice
	var err error

	// Retry loop for getting account balances
	for i := 0; i < maxRetries; i++ {
		account, err = t.Client.NewGetAccountService().Do(context.Background())
		if err == nil {
			break
		}
		log.Printf("Error getting account balances: %v", err)
		time.Sleep(retryDelay)
	}
	if err != nil {
		log.Fatalf("Failed to get account balances after %d attempts: %v", maxRetries, err)
		return nil
	}

	// Retry loop for getting prices
	for i := 0; i < maxRetries; i++ {
		prices, err = t.Client.NewListPricesService().Do(context.Background())
		if err == nil {
			break
		}
		log.Printf("Error getting prices: %v", err)
		time.Sleep(retryDelay)
	}
	if err != nil {
		log.Fatalf("Failed to get prices after %d attempts: %v", maxRetries, err)
		return nil
	}

	// Create a map to store the prices for easy lookup
	priceMap := make(map[string]float64)
	for _, p := range prices {
		price, err := strconv.ParseFloat(p.Price, 64)
		if err != nil {
			log.Printf("Error parsing price for %s: %v", p.Symbol, err)
			continue
		}
		priceMap[p.Symbol] = price
	}

	balances := map[string]string{}
	for _, balance := range account.Balances {
		assetBalance, _ := decimal.NewFromString(balance.Free)
		if balance.Free != "0" && !strings.HasPrefix(balance.Asset, "LD") {
			usdtValue := assetBalance
			if balance.Asset != "USDT" {
				price, ok := priceMap[balance.Asset+"USDT"]
				if !ok {
					log.Printf("Price not found for %sUSDT", balance.Asset)
					continue
				}
				usdtValue = assetBalance.Mul(decimal.NewFromFloat(price))
			}
			if usdtValue.GreaterThan(decimal.NewFromFloat(0.5)) {
				log.Printf("%s: %s, %s\n", balance.Asset, balance.Free, usdtValue.String())
				balances[balance.Asset] = balance.Free
			}
		}
	}
	return balances
}

func (t *BinanceTrader) GetPrices() map[string]float64 {
	var prices []*binance.SymbolPrice

	const retryDelay = 2 * time.Second
	var err error
	for i := 0; i < maxRetries; i++ {
		prices, err = t.Client.NewListPricesService().Do(context.Background())
		if err == nil {
			break
		}
		log.Printf("Error getting prices: %v", err)
		time.Sleep(retryDelay)
	}
	if err != nil {
		log.Fatalf("Failed to get prices after %d attempts: %v", maxRetries, err)
		return nil
	}

	// Create a map to store the prices for easy lookup
	priceMap := make(map[string]float64)
	for _, p := range prices {
		price, err := strconv.ParseFloat(p.Price, 64)
		if err != nil {
			log.Printf("Error parsing price for %s: %v", p.Symbol, err)
			continue
		}
		priceMap[p.Symbol] = price
	}
	return priceMap
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

	// Fetch trading rules
	exchangeInfo, err := t.Client.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		log.Fatalf("Error getting exchange info: %v", err)
	}

	var stepSize, minQty, maxQty float64
	for _, symbolInfo := range exchangeInfo.Symbols {
		if symbolInfo.Symbol == symbol {
			for _, filter := range symbolInfo.Filters {
				if filter["filterType"].(string) == "LOT_SIZE" {
					stepSize, _ = strconv.ParseFloat(filter["stepSize"].(string), 64)
					minQty, _ = strconv.ParseFloat(filter["minQty"].(string), 64)
					maxQty, _ = strconv.ParseFloat(filter["maxQty"].(string), 64)
					break
				}
			}
			break
		}
	}

	// Ensure quantity meets the constraints
	quantity = math.Floor(quantity/stepSize) * stepSize
	if quantity < minQty {
		quantity = minQty
	} else if quantity > maxQty {
		quantity = maxQty
	}

	log.Printf("Price %s: %f\n", symbol, price)
	log.Printf("We are about to place %f %s for %f %s\n", quantity, base, amount, counter)

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
