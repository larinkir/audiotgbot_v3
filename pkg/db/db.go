package db

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

// Подключение к ДБ
func ConnectToDb(dataSourceName string) *sql.DB {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// Поиск книг в ДБ
func searchBookInDb(formatRequst string, db *sql.DB) *sql.Rows {
	rows, err := db.Query("SELECT * FROM books WHERE to_tsvector('russian', author) @@ to_tsquery('russian', $1) OR to_tsvector('russian', book_name) @@ to_tsquery('russian', $1) ", formatRequst)
	if err != nil {
		log.Fatal("Ошибка в функции searchBookInDb:", err)
	}
	return rows
}

// Формирование списка книг из БД
func MakeListBooks(requestBook string, dbConnect *sql.DB) ([]string, int) {

	type Book struct {
		book_name string
		author    string
		reader    string
		file_id   string
		book_id   string
	}

	//Количество найденных книг
	var numBook int

	//Сформированный список книг
	var listBooks []string

	formatRequst := strings.Replace(requestBook, " ", "&", -1)

	rows := searchBookInDb(formatRequst, dbConnect)

	defer rows.Close()

	for rows.Next() {

		bk := new(Book)
		err := rows.Scan(&bk.book_id, &bk.book_name, &bk.author, &bk.reader, &bk.file_id)
		if err != nil {
			log.Fatal("Ошибка в функции MakeListBooks. Не удалось распарсить данные в структуру Book. ", err)
		}

		book := fmt.Sprintf("📖 <b>%s.</b>\n  Автор: %s | Читает: %s\n %s\n\n", bk.book_name, bk.author, bk.reader, bk.book_id)
		listBooks = append(listBooks, book)
		numBook++
	}

	return listBooks, numBook
}

// Поиск файла книги в БД
func SearchFileBook(book_id string, dbConnect *sql.DB) string {

	var file_id string

	err := dbConnect.QueryRow("SELECT file_id FROM books WHERE book_id = $1", book_id).Scan(&file_id)
	if err != nil {
		log.Fatal("Ошибка в функции SearchFileBook:", err)
	}
	return file_id
}
