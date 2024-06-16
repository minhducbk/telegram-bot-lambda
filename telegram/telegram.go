package telegram

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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

func SendToTelegramChannel(message string) {
	bot := initBot()
	c := telegramChannel(bot)
	// getChannelIDs(bot)

	message = fmt.Sprintf("[UPDATE] At %s, hi teacher @ducbk95\n Someone is calling to this API with these info:\n %s", time.Now().Format(time.RFC3339), message)
	bot.Send(tgbotapi.NewMessage(c.ID, message))
}
