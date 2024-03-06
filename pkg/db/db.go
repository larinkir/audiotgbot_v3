package db

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

type Book struct {
	book_id   int
	book_name string
	author    string
	reader    string
}

// Подключение к ДБ
func ConnectToDb(dataSourceName string) *sql.DB {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// Поиск книг в ДБ
func searchBookInDb(formatRequest string, db *sql.DB) *sql.Rows {

	rows, err := db.Query("SELECT books.id, books.name, authors.name, performers.name "+
		"FROM books "+
		"join authors on books.author_id = authors.id "+
		"join performers on books.performer_id = performers.id "+
		"WHERE to_tsvector('russian', authors.name) @@ to_tsquery('russian', $1) OR to_tsvector('russian', books.name) @@ to_tsquery('russian', $1) ", formatRequest)
	if err != nil {
		log.Fatal("Ошибка в функции searchBookInDb:", err)
	}
	fmt.Println("Строка rows:", rows)
	return rows
}

// Формирование списка книг из БД
func MakeListBooks(requestBook string, dbConnect *sql.DB) ([]string, int) {

	//Сформированный список книг
	var listBooks []string

	formatRequest := strings.Replace(requestBook, " ", "&", -1)

	rows := searchBookInDb(formatRequest, dbConnect)

	defer rows.Close()

	for rows.Next() {

		bk := new(Book)
		err := rows.Scan(&bk.book_id, &bk.book_name, &bk.author, &bk.reader)
		if err != nil {
			log.Fatal("Ошибка в функции MakeListBooks. Не удалось распарсить данные в структуру Book. ", err)
		}

		book := fmt.Sprintf("📖 <b>%s.</b>\n  Автор: %s | Читает: %s\n /book%d\n\n", bk.book_name, bk.author, bk.reader, bk.book_id)
		listBooks = append(listBooks, book)

	}

	return listBooks, len(listBooks)
}

// Поиск файла книги в БД
func SearchFileBook(book_id string, dbConnect *sql.DB) string {

	//Поменять название  file_id
	var file_id string

	err := dbConnect.QueryRow("SELECT file_id FROM books WHERE id = $1", book_id).Scan(&file_id)
	if err != nil {
		log.Fatal("Ошибка в функции SearchFileBook:", err)
	}
	return file_id
}
