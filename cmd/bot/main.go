package main

import (
	"audio_tg_bot_v3/pkg/services"
	"audio_tg_bot_v3/pkg/telegram"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {

	//Получение токена Бота
	tgbottoken := services.GetToken("cnfg.env")

	//Создание объекта Бота
	bot, err := tgbotapi.NewBotAPI(tgbottoken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	telegram.WorkBot(bot)

}
