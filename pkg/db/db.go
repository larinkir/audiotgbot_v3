package db

import (
	"audio_tg_bot_v3/pkg/services"
	"database/sql"
	"fmt"
	"log"
	"strings"
)

// Формирование списка книг из БД
func MakeListBooks(requestBook string) ([]string, int) {

	type Book struct {
		book_name string
		author    string
		reader    string
		file_id   string
		book_id   string
	}

	dataSourceName := services.GetKeyDb("cnfg.env")

	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	formatRequst := strings.Replace(requestBook, " ", "&", -1)

	rows, err := db.Query("SELECT * FROM books WHERE to_tsvector('russian', author) @@ to_tsquery('russian', $1) OR to_tsvector('russian', book_name) @@ to_tsquery('russian', $1) ", formatRequst)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var numBook int
	var listBooks []string

	for rows.Next() {

		bk := new(Book)
		err := rows.Scan(&bk.book_id, &bk.book_name, &bk.author, &bk.reader, &bk.file_id)
		if err != nil {
			log.Fatal(err)
		}

		book := fmt.Sprintf("📖 <b>%s.</b>\n  Автор: %s | Читает: %s\n %s\n\n", bk.book_name, bk.author, bk.reader, bk.book_id)
		listBooks = append(listBooks, book)
		numBook++

	}

	return listBooks, numBook
}

// Поиск файла книги в БД
func SearchFileBook(book_id string) string {

	dataSourceName := services.GetKeyDb("cnfg.env")

	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var file_id string

	err = db.QueryRow("SELECT file_id FROM books WHERE book_id = $1", book_id).Scan(&file_id)
	if err != nil {
		log.Fatal(err)
	}

	return file_id

}
