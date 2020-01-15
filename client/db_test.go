package main

import (
	"os"
	"testing"
	"time"
)

func TestDB(t *testing.T) {
	// db open
	if err := initDB("test.db"); err != nil {
		t.Fatal(err)
	}
	defer closeDB()

	// query 0
	if exist, err := existData("file0001", nil); err != nil {
		t.Error("query data error", err)
	} else {
		if exist {
			t.Error("not empty key")
		}
	}

	// add new
	if err := addData("file0001", []byte("testdbdbdbdbaaa"), time.Now()); err != nil {
		t.Fatal(err)
	}

	// query 1
	if exist, err := existData("file0001", nil); err != nil {
		t.Error("query data error", err)
	} else {
		if !exist {
			t.Error("not exist key")
		}
	}

	// del
	if err := delData("file0001"); err != nil {
		t.Error(err)
	}

	// query 2
	if exist, err := existData("file0001", nil); err != nil {
		t.Error("query data error", err)
	} else {
		if exist {
			t.Error("exist key")
		}
	}
}

func TestFinal(t *testing.T) {
	os.Remove("test.db")
}
