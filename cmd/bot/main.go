package main

import (
	"audio_tg_bot_v3/pkg/db"
	"audio_tg_bot_v3/pkg/services"
	"audio_tg_bot_v3/pkg/telegram"

	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {

	//Получение токена Бота
	tgbottoken := services.GetToken("cnfg.env")

	//Получение ключа ДБ
	dataSourceName := services.GetKeyDb("cnfg.env")
	//Подключение к ДБ
	dbConnect := db.ConnectToDb(dataSourceName)
	defer dbConnect.Close()

	//Создание объекта Бота
	bot, err := tgbotapi.NewBotAPI(tgbottoken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	telegram.WorkBot(bot, dbConnect)

}
