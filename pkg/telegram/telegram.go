package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	binance "github.com/adshao/go-binance/v2"
	"github.com/shopspring/decimal"

	"telegram-bot/pkg/ip"
	"telegram-bot/pkg/trader"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const OrderStatusFilled = "FILLED"

type Alert struct {
	Symbol string `json:"symbol"`
	Label  string `json:"label"`
}

const (
	TradeSize      float64 = 30 // 30 USDT
	CandleInterval         = 30 * time.Minute
)

var AvailableLabels = []string{"Buy", "Wave 3 Start", "Wave 3 End", "Wave 2 Start", "Wave 4 Start", "Wave A Start", "Wave C Start"}

func botToken() string {
	return os.Getenv("BOT_TOKEN")
}

func teleAutoTradingID() int64 {
	id, _ := strconv.ParseInt(os.Getenv("TELE_GROUP_ID"), 10, 64)
	return id
}

func initBot() *tgbotapi.BotAPI {
	bot, err := tgbotapi.NewBotAPI(botToken())
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s\n", bot.Self.UserName)
	return bot
}

func getChannelIDs(bot *tgbotapi.BotAPI) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	channelIDs := make(map[int64]bool)
	for update := range updates {
		if update.Message != nil && update.Message.Chat.IsChannel() {
			channelIDs[update.Message.Chat.ID] = true
		}
	}

	log.Println("Available Channel IDs:")
	for id := range channelIDs {
		fmt.Println(id)
	}
}

func telegramChannel(bot *tgbotapi.BotAPI) tgbotapi.Chat {
	c, err := bot.GetChat(tgbotapi.ChatInfoConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: teleAutoTradingID(),
		},
	})
	if err != nil {
		log.Panic(err)
	}
	return c
}

func sendMessageWithRetry(bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig, retries int) {
	for i := 0; i < retries; i++ {
		_, err := bot.Send(msg)
		if err == nil {
			return
		}
		log.Printf("Error sending message to Telegram: %v. Retrying %d/%d...\n", err, i+1, retries)
		time.Sleep(time.Duration(2^i) * time.Second) // Exponential backoff
	}
	log.Printf("Failed to send message to Telegram after %d retries\n", retries)
}

func SendToTelegramChannel(eventBody string) {
	var alert Alert

	// Parse the JSON string into the alert variable
	err := json.Unmarshal([]byte(eventBody), &alert)
	if err != nil {
		log.Fatalf("Error parsing JSON: %v, eventBody: %s", err, eventBody)
	}
	alertCoin := alert.Symbol
	suffix := "USDT"

	if strings.HasSuffix(alertCoin, suffix) {
		alertCoin = alertCoin[:len(alertCoin)-len(suffix)]
	}

	log.Println("before initBot")
	bot := initBot()
	log.Println("after initBot")
	c := telegramChannel(bot)

	// Send Lambda function IP
	ip, err := ip.GetPublicIP()
	if err == nil {
		msg := tgbotapi.NewMessage(c.ID, "Lambda Function IP: "+ip)
		sendMessageWithRetry(bot, msg, 10)
	}

	// Format and send the message
	newMessage := fmt.Sprintf("[UPDATE] At %s, @ducbk95\n Someone is calling to this API with this body:\n %s", time.Now().Format(time.RFC3339), eventBody)
	msg := tgbotapi.NewMessage(c.ID, newMessage)
	sendMessageWithRetry(bot, msg, 10)

	// Fetch Binance balances
	trader := trader.NewBinanceTrader()

	beforeBalanceMessage := fmt.Sprintf("<Before> Balances: \n")
	balances := trader.GetBalances()
	for asset, balance := range balances {
		beforeBalanceMessage += fmt.Sprintf("%s %s\n", balance, asset)
	}
	msg = tgbotapi.NewMessage(c.ID, beforeBalanceMessage)
	sendMessageWithRetry(bot, msg, 10)

	log.Println("Processing: ", alert, alertCoin)
	// Process alert based on rules
	processAlert(trader, alert, alertCoin)

	// Fetch Binance balances again
	otherMessage := fmt.Sprintf("<After> Balances: \n")
	balances = trader.GetBalances()
	for asset, balance := range balances {
		otherMessage += fmt.Sprintf("%s %s\n", balance, asset)
	}
	msg = tgbotapi.NewMessage(c.ID, otherMessage)
	sendMessageWithRetry(bot, msg, 10)
}

func processAlert(trader *trader.BinanceTrader, alert Alert, alertCoin string) {
	usdtBalance, _ := decimal.NewFromString(trader.GetBalances()["USDT"])
	tradeSize := decimal.NewFromFloat(30)
	prices := trader.GetPrices()
	coinBalance, _ := decimal.NewFromString(trader.GetBalances()[alertCoin])
	coinBalanceInFloat, _ := coinBalance.Float64()
	fmt.Println("Coin balance: ", coinBalanceInFloat, ". Coin price: ", prices[alert.Symbol])
	switch alert.Label {
	case "Buy", "Wave 3 Start":
		// Check if there is an existing trade
		if coinBalanceInFloat*prices[alert.Symbol] > 1 {
			log.Printf("Existing trade for %s, skipping buy order", alertCoin)
			return
		}

		// Determine the size of the trade
		if usdtBalance.GreaterThan(tradeSize) {
			tradeSize = decimal.NewFromFloat(30)
		} else {
			tradeSize = usdtBalance
		}

		if tradeSize.GreaterThan(decimal.NewFromFloat(0)) {
			tradeSizeFloat, _ := tradeSize.Float64()
			trader.PlaceMarketBuyOrder(alertCoin, "USDT", tradeSizeFloat)
		} else {
			log.Println("Insufficient USDT balance to place a buy order")
		}
	case "Wave 3 End", "Wave 2 Start", "Wave 4 Start", "Wave A Start", "Wave C Start":
		// Check if there is an existing trade

		if trade, exists := getTrade(trader, alert.Symbol); exists && (coinBalanceInFloat*prices[alert.Symbol] > 0.05) {
			// Check if the exit signal is within the same candle interval as the entry
			if time.Since(time.Unix(trade.Time/1000, 0)) < CandleInterval {
				log.Printf("Exit signal for %s occurred within the same candle interval, skipping sell order", alert.Symbol)
				return
			}

			quantity, _ := decimal.NewFromString(trader.GetBalances()[alertCoin])
			trader.PlaceMarketSellOrder(alertCoin, "USDT", quantity.InexactFloat64())
		} else {
			log.Printf("No existing trade for %s, skipping sell order", alertCoin)
		}
	default:
		log.Printf("Unknown label: %s", alert.Label)
	}
}

func isExistingTrade(trader *trader.BinanceTrader, symbol string) bool {
	orders, err := trader.Client.NewListOrdersService().Symbol(symbol).Do(context.Background())
	if err != nil {
		log.Fatalf("Error fetching order history for %s: %v", symbol, err)
	}
	for _, order := range orders {
		if order.Status == binance.OrderStatusTypeFilled && order.Side == binance.SideTypeBuy {
			return true
		}
	}
	return false
}

func getTrade(trader *trader.BinanceTrader, symbol string) (binance.Order, bool) {
	orders, err := trader.Client.NewListOrdersService().Symbol(symbol).Do(context.Background())
	if err != nil {
		log.Fatalf("Error fetching order history for %s: %v", symbol, err)
	}

	// Sort orders by creation time in descending order
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].Time > orders[j].Time
	})

	for _, order := range orders {
		if order.Status == binance.OrderStatusTypeFilled && order.Side == binance.SideTypeBuy {
			return *order, true
		}
	}
	return binance.Order{}, false
}

// {
// "symbol": "{{ticker}}",
// "label": "Buy"
// }
