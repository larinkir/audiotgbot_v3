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
		log.Fatal("Ошибка в telegram.go -> GetUpdatesChan:", err)
	}

	for update := range updates {
		if update.Message != nil {

			// Обработка команд
			if update.Message.IsCommand() {
				err := handlerCommand(update, bot, dbConnect)
				if err != nil {
					log.Println("Ошибка в telegram.go - > handlerCommand", err)
				}
				continue
			}

			//Обработка файлов
			if isAudioFile(update.Message) {
				handlerFile(update.Message.Audio)
				continue
			}

			//Обработка запроса пользователя
			err := handlerRequests(update, bot, dbConnect)
			if err != nil {
				log.Println("Ошибка в telegram.go -> handlerRequests", err)
			}

			//Обработка обновлений с инлайн клваиатуры
		} else if update.CallbackQuery != nil {
			err := handlerUpdKeyboard(update, bot, dbConnect)
			if err != nil {
				log.Println("Ошибка в telegram.go -> handlerUpdKeyboard", err)
			}

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
func handlerCommand(update tgbotapi.Update, bot *tgbotapi.BotAPI, dbConnect *sql.DB) error {
	switch {
	case update.Message.Text == "/start":
		_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Бот запущен."))
		if err != nil {
			return err
		}
		_, err = bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Для поиска книги, напишите название в сообщении."))
		if err != nil {
			return err
		}

	case update.Message.Text[:5] == "/book":
		fileId := db.SearchFileBook(update.Message.Text[5:], dbConnect)
		_, err := bot.Send(tgbotapi.NewAudioShare(update.Message.Chat.ID, fileId))
		if err != nil {
			return err
		}

	default:
		_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда."))
		if err != nil {
			return err
		}
	}

	return nil

}

func handlerRequests(update tgbotapi.Update, bot *tgbotapi.BotAPI, dbConnect *sql.DB) error {

	//Добавление запроса с Id в историю запросов
	err := services.AddRequest(update.Message.MessageID, update.Message.Text)
	if err != nil {
		log.Println("Ошибка в handlerRequests -> AddRequest:", err)
	}

	//Формирование списка книг из БД
	listBooks, numBook := db.MakeListBooks(update.Message.Text, dbConnect)

	var books string

	switch {

	//Если книг не найдено
	case numBook == 0:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Такую книгу найти не удалось.")
		_, err := bot.Send(msg)
		if err != nil {
			return err
		}

		//Если количество книг <= 5
	case numBook <= 5:
		for i := 0; i < numBook; i++ {
			books += listBooks[i]
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, books)
		msg.ParseMode = "HTML"
		_, err = bot.Send(msg)
		if err != nil {
			return err
		}

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
		_, err = bot.Send(msg)
		if err != nil {
			return err
		}

	}

	return nil

}

func handlerUpdKeyboard(update tgbotapi.Update, bot *tgbotapi.BotAPI, dbConnect *sql.DB) error {

	//Обработать ошибку!
	updateButton, err := strconv.Atoi(update.CallbackQuery.Data)
	if err != nil {
		log.Println("Ошибка при чтении update в функции -> handlerUpdKeyboard ")
		return err
	}

	// Id обновления с инлайн клавиатуры
	msgId := update.CallbackQuery.Message.MessageID

	//Извлечение запроса из истори запросов по ID
	request, err := services.SearchRequstId(msgId - 1)
	if err != nil {
		log.Println("Ошибка в функции SearchRequestId:", err)
		return err
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
	numButtons, err := services.SearchKeyboardId(msgId - 1)
	if err != nil {
		log.Println("Ошибка в функции SearchKeyboardId:", err)
	}

	//Вызов инлайн клавиатуры
	numericKeyboard := makeButtonsNext(numButtons, updateButton)

	//Обновление сообщения
	msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, msgId, books)
	msg.ReplyMarkup = &numericKeyboard
	msg.ParseMode = "HTML"
	_, err = bot.Send(msg)
	if err != nil {
		return err
	}
	return nil
}

// Проверка присутствия в сообщении audio файла.
func isAudioFile(m *tgbotapi.Message) bool {
	if m.Audio != nil {
		return true
	}
	return false
}

func handlerFile(audio *tgbotapi.Audio) {

	title := audio.Title
	performer := audio.Performer
	fileId := audio.FileID

	err := services.AddBooks(title, performer, fileId)
	if err != nil {
		log.Println("Ошибка в handlerFile -> AddBooks", err)
	}
}
