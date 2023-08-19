package telegram

import (
	"audio_tg_bot_v3/pkg/db"
	"audio_tg_bot_v3/pkg/services"
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/lib/pq"
)

func WorkBot(bot *tgbotapi.BotAPI) {

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}

	for update := range updates {
		if update.Message != nil {
			// Обработка команд
			if update.Message.IsCommand() {
				switch {
				case update.Message.Text == "/start":
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Бот запущен."))
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Для поиска книги, напишите название в сообщении."))

				case update.Message.Text[:2] == "/b":
					file_id := db.SearchFileBook(update.Message.Text)
					bot.Send(tgbotapi.NewAudioShare(update.Message.Chat.ID, file_id))

				default:
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда."))
				}
				continue
			}
			//Обработка файлов.
			if update.Message.Audio != nil {
				title := update.Message.Audio.Title
				author_reader := update.Message.Audio.Performer
				file_id := update.Message.Audio.FileID

				err := services.AddBooks(title, author_reader, file_id)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Книгу добавить не удалось"))
				} else {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Книга "+title+author_reader+" успешно добавлена"))
				}

				continue
			}

			//Запрос пользователя с Id
			request := []string{strconv.Itoa(update.Message.MessageID), update.Message.Text}

			//Добавление запроса с Id в базу
			services.AddRequest(request)

			//Формирование списка книг из БД
			listBooks, numBook := db.MakeListBooks(update.Message.Text)

			var books string

			switch {
			case numBook == 0:
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Такую книгу найти не удалось.")
				bot.Send(msg)

			case numBook <= 5:
				for i := 0; i < numBook; i++ {
					books += listBooks[i]
				}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, books)
				msg.ParseMode = "HTML"
				bot.Send(msg)

			case numBook > 5:
				for i := 0; i < 5; i++ {
					books += listBooks[i]
				}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, books)
				msg.ParseMode = "HTML"

				numButtons := numBook/5 + 1

				//Сохранение количества кнопок для каждого Id запроса
				inlineKeyboard := []string{strconv.Itoa(update.Message.MessageID), strconv.Itoa(numButtons)}
				services.AddInlineKeyboard(inlineKeyboard)

				//Вызов инлайн клавиатуры
				numericKeyboard := makeButtonsFirst(numButtons)
				msg.ReplyMarkup = numericKeyboard
				bot.Send(msg)

			}

		} else if update.CallbackQuery != nil {

			updateButton, _ := strconv.Atoi(update.CallbackQuery.Data)

			// Id обновления с инлайн клавиатуры
			msgId := update.CallbackQuery.Message.MessageID

			//Извлечение запроса из списка запросов по ID
			request := services.SearchRequstId(msgId - 1)

			//Формирование списка книг из БД
			listBooks, numBook := db.MakeListBooks(request)

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

			//Вызов инлайн клавиатуры
			numButtons := services.SeachKeyboardId(msgId - 1)
			numericKeyboard := makeButtonsNext(numButtons, updateButton)

			msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, msgId, books)
			msg.ReplyMarkup = &numericKeyboard
			msg.ParseMode = "HTML"

			bot.Send(msg)

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
