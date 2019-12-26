package main

import (
	"database/sql"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var lo sync.RWMutex

func initDB(path string) error {
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

func addData(path string, md5 string, date time.Time) error {
	lo.Lock()
	defer lo.Unlock()

	statement, _ := db.Prepare("INSERT INTO files (pid, md5,regdate) VALUES (?, ?, ?)")
	if _, err := statement.Exec(path, md5, date); err != nil {
		return err
	}
	return nil
}

func existData(path string) (bool, error) {
	lo.RLock()
	defer lo.RUnlock()

	var exist int
	rows, _ := db.Query("SELECT EXISTS(SELECT * FROM files WHERE pid = ?)", path)

	for rows.Next() {
		rows.Scan(&exist)
	}

	return (exist == 1), nil
}

func delData(path string) error {
	lo.Lock()
	defer lo.Unlock()

	statement, _ := db.Prepare("DELETE FROM files WHERE pid = ?")
	if _, err := statement.Exec(path); err != nil {
		return err
	}
	return nil
}
