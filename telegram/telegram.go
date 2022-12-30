package telegram

import (
	"fmt"
	"log"
	"os"
	"telegram-bot/price"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)


const BotToken = "5869911418:AAF0KoALqB--cnqktty-Gn8LwqGbx1s3AF8"
const TeleChannelID = -860049871

func botToken() string {
	if os.Getenv("BOT_TOKEN") == "" {
		return BotToken
	}
	return os.Getenv("BOT_TOKEN")
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

func telegramChannel(bot *tgbotapi.BotAPI) tgbotapi.Chat {
	c, err := bot.GetChat(tgbotapi.ChatInfoConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: TeleChannelID,
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
	if message == "" {
		message = fmt.Sprintf("[UPDATE] At %s, hi teachers @phuocleanh @ldt25290 @thieunv @ducbk95\n %s", time.Now().Format(time.RFC3339), price.PricesMessage())
	}
	teleMsg := tgbotapi.NewMessage(c.ID, message)
	bot.Send(teleMsg)
}
