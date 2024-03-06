package services

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Получение токена Бота
func GetToken(configName string) string {
	err := godotenv.Load(configName)
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	tgBotToken := os.Getenv("TG_BOT_TOKEN")

	return tgBotToken
}

// Получение данных ДБ
func GetKeyDb(configName string) string {

	err := godotenv.Load(configName)
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	loginDb := os.Getenv("loginDb")
	passwordDb := os.Getenv("passwordDb")
	nameDb := os.Getenv("nameDb")

	dataSourceName := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", loginDb, passwordDb, nameDb)

	return dataSourceName

}

// Сохранение запроса от пользователя вместе с Id
func AddRequest(msgId int, textMessage string) error {

	request := []string{strconv.Itoa(msgId), textMessage}
	file, err := os.OpenFile("dataRequestsId.csv", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	err = writer.Write(request)
	if err != nil {
		return err
	}
	defer writer.Flush()

	return nil

}

// Поиск запроса по id
func SearchRequstId(msgId int) (string, error) {
	//Перенести переменную перед использованием (перенести на 78 строку)
	msgIdstr := strconv.Itoa(msgId)

	file, err := os.Open("dataRequestsId.csv")
	if err != nil {
		return "", err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return "", err
	}

	for _, row := range records {
		if row[0] == msgIdstr {
			return row[1], nil
		}
	}

	return "", nil
}

// Сохранение количества страниц для инлайн клавиатуры
func AddInlineKeyboard(msgId, numButtons int) error {

	inlineKeyboard := []string{strconv.Itoa(msgId), strconv.Itoa(numButtons)}

	file, err := os.OpenFile("dataInlineId.csv", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	err = writer.Write(inlineKeyboard)
	if err != nil {
		return err
	}
	defer writer.Flush()

	return nil

}

func SearchKeyboardId(msgId int) (int, error) {

	msgIdstr := strconv.Itoa(msgId)

	file, err := os.Open("dataInlineId.csv")
	if err != nil {
		return 0, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return 0, err
	}

	for _, row := range records {
		if row[0] == msgIdstr {
			numButtons, _ := strconv.Atoi(row[1])
			return numButtons, nil
		}
	}

	return 0, nil

}

func AddBooks(title, author, file_id string) error {
	file, err := os.OpenFile("books_db.csv", os.O_WRONLY|os.O_APPEND, 0644)

	if err != nil {
		return err
	}

	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Comma = ';'
	defer writer.Flush()

	err = writer.Write([]string{title, author, file_id})
	if err != nil {
		return err
	}

	return nil
}
