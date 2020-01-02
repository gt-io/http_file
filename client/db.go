package main

/* use bolt db  */
import (
	"bytes"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

var db *bolt.DB
var lo sync.RWMutex

const myBucket = "files"

func initDB(path string) error {
	lo.Lock()
	defer lo.Unlock()

	database, err := bolt.Open(path, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	if err := database.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(myBucket))
		if err != nil {
			return fmt.Errorf("Create bucker: %s", err)
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	db = database

	return nil
}

func closeDB() {
	lo.Lock()
	defer lo.Unlock()

	if db != nil {
		db.Close()
	}
}

func addData(path string, md5 []byte, date time.Time) error {
	lo.Lock()
	defer lo.Unlock()

	return db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(myBucket)).Put([]byte(path), []byte(md5))
	})
}

func existData(path string, md5 []byte) (bool, error) {
	lo.RLock()
	defer lo.RUnlock()

	result := true
	db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket([]byte(myBucket)).Get([]byte(path))
		if v == nil || len(v) == 0 || bytes.Compare(v, md5) != 0 {
			result = false
		}
		return nil
	})

	return result, nil
}

func delData(path string) error {
	lo.Lock()
	defer lo.Unlock()

	return db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(myBucket)).Put([]byte(path), nil)
	})
}

/* use sqllite3
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
*/
