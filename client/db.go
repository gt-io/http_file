package main

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitDB(path string) error {
	database, err := sql.Open("sqlite3", path)
	if err != nil {
		return err
	}

	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS files (pid TEXT PRIMARY KEY, md5 TEXT, regdate DATETIME)")
	if _, err := statement.Exec(); err != nil {
		return err
	}

	db = database
	return nil
}

func AddData(path string, md5 string, date time.Time) error {
	statement, _ := db.Prepare("INSERT INTO files (pid, md5,regdate) VALUES (?, ?, ?)")
	if _, err := statement.Exec(path, md5, date); err != nil {
		return err
	}
	return nil
}

func ExistData(path string) (bool, error) {
	var exist int

	rows, _ := db.Query("SELECT EXISTS(SELECT * FROM files WHERE pid = ?)", path)

	for rows.Next() {
		rows.Scan(&exist)
	}

	return (exist == 1), nil
}

func DelData(path string) error {

	statement, _ := db.Prepare("DELETE FROM files WHERE pid = ?")
	if _, err := statement.Exec(path); err != nil {
		return err
	}
	return nil
}
