package telegram

import (
	"audio_tg_bot_v3/pkg/db"
	"audio_tg_bot_v3/pkg/services"
	"database/sql"
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/lib/pq"
)

func WorkBot(bot *tgbotapi.BotAPI, dbConnect *sql.DB) {

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal("Ошибка при получении апдейтов:", err)
	}

	for update := range updates {
		if update.Message != nil {

			// Обработка команд
			if update.Message.IsCommand() {
				handlerCommand(update, bot, dbConnect)
				continue
			}

			//Обработка запроса пользователя
			handlerRequsts(update, bot, dbConnect)

			//Обработка обновлений с инлайн клваиатуры
		} else if update.CallbackQuery != nil {
			handlerUpdKeyboard(update, bot, dbConnect)

		}

	}

}

// Первый вызов инлайн клавиатуры
func makeButtonsFirst(numButtons int) tgbotapi.InlineKeyboardMarkup {
	num := strconv.Itoa(numButtons)
	keyboardRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("1/"+num, "1"),
		tgbotapi.NewInlineKeyboardButtonData("➡️", "2")}

	return tgbotapi.NewInlineKeyboardMarkup(keyboardRow)
}

// Последующие вызовы инлайн клавиатуры
func makeButtonsNext(numButtons, updateButton int) tgbotapi.InlineKeyboardMarkup {
	nb := strconv.Itoa(numButtons)
	ub := strconv.Itoa(updateButton)
	leftButton := strconv.Itoa(updateButton - 1)
	rightButton := strconv.Itoa(updateButton + 1)

	switch updateButton {
	case 1:
		keyboardRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("1/"+nb, "1"),
			tgbotapi.NewInlineKeyboardButtonData("➡️", "2")}

		return tgbotapi.NewInlineKeyboardMarkup(keyboardRow)

	case numButtons:
		keyboardRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("⬅️", leftButton),
			tgbotapi.NewInlineKeyboardButtonData(ub+"/"+nb, ub)}

		return tgbotapi.NewInlineKeyboardMarkup(keyboardRow)

	default:
		keyboardRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("⬅️", leftButton),
			tgbotapi.NewInlineKeyboardButtonData(ub+"/"+nb, ub),
			tgbotapi.NewInlineKeyboardButtonData("➡️", rightButton)}

		return tgbotapi.NewInlineKeyboardMarkup(keyboardRow)
	}

}

// Обработчик команд
func handlerCommand(update tgbotapi.Update, bot *tgbotapi.BotAPI, dbConnect *sql.DB) {
	switch {
	case update.Message.Text == "/start":
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Бот запущен."))
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Для поиска книги, напишите название в сообщении."))

	case update.Message.Text[:2] == "/b":
		file_id := db.SearchFileBook(update.Message.Text, dbConnect)
		bot.Send(tgbotapi.NewAudioShare(update.Message.Chat.ID, file_id))

	default:
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда."))
	}

}

func handlerRequsts(update tgbotapi.Update, bot *tgbotapi.BotAPI, dbConnect *sql.DB) {

	//Добавление запроса с Id в историю запросов
	err := services.AddRequest(update.Message.MessageID, update.Message.Text)
	if err != nil {
		log.Println("Ошибка в функции AddRequest:", err)
	}

	//Формирование списка книг из БД
	listBooks, numBook := db.MakeListBooks(update.Message.Text, dbConnect)

	var books string

	switch {

	//Если книг не найдено
	case numBook == 0:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Такую книгу найти не удалось.")
		bot.Send(msg)

		//Если количество книг <= 5
	case numBook <= 5:
		for i := 0; i < numBook; i++ {
			books += listBooks[i]
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, books)
		msg.ParseMode = "HTML"
		bot.Send(msg)

		//Если количество книг >5 вызывается инлайн клавиатура
	case numBook > 5:
		for i := 0; i < 5; i++ {
			books += listBooks[i]
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, books)
		msg.ParseMode = "HTML"

		var numButtons int
		if numBook%5 == 0 {
			numButtons = numBook / 5
		} else {
			numButtons = numBook/5 + 1
		}

		//Сохранение количества кнопок для каждого Id запроса
		err := services.AddInlineKeyboard(update.Message.MessageID, numButtons)
		if err != nil {
			log.Println("Ошибка в функции AddInlineKeyboard:", err)
		}

		//Вызов инлайн клавиатуры
		numericKeyboard := makeButtonsFirst(numButtons)
		msg.ReplyMarkup = numericKeyboard
		bot.Send(msg)

	}

}

func handlerUpdKeyboard(update tgbotapi.Update, bot *tgbotapi.BotAPI, dbConnect *sql.DB) {

	//Обработать ошибку!
	updateButton, _ := strconv.Atoi(update.CallbackQuery.Data)

	// Id обновления с инлайн клавиатуры
	msgId := update.CallbackQuery.Message.MessageID

	//Извлечение запроса из истори запросов по ID
	request, err := services.SearchRequstId(msgId - 1)
	if err != nil {
		log.Println("Ошибка в функции SearchRequstId:", err)
	}

	//Формирование списка книг из БД
	listBooks, numBook := db.MakeListBooks(request, dbConnect)

	var books string

	if numBook > updateButton*5 {
		for i := 5 * (updateButton - 1); i < 5*updateButton; i++ {
			books += listBooks[i]
		}

	} else {
		for i := 5 * (updateButton - 1); i < numBook; i++ {
			books += listBooks[i]
		}
	}

	//Поиск количества кнопок по Id сообщения
	numButtons, err := services.SeachKeyboardId(msgId - 1)
	if err != nil {
		log.Println("Ошибка в функции SeachKeyboardId:", err)
	}

	//Вызов инлайн клавиатуры
	numericKeyboard := makeButtonsNext(numButtons, updateButton)

	//Обновление сообщения
	msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, msgId, books)
	msg.ReplyMarkup = &numericKeyboard
	msg.ParseMode = "HTML"
	bot.Send(msg)

}
