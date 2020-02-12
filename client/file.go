package main

import (
	"errors"
	"log"
	"math"
	"os"
	"time"
)

func openFile(filePath string, wait time.Duration) (*os.File, error) {
	fi, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	if fi.IsDir() {
		return nil, errors.New("path is dir")
	}

	retry := math.Ceil(wait.Seconds() / 3)
	for retry > 0 {
		if time.Since(fi.ModTime()) < time.Second*10 {
			log.Println("file is busy", filePath, retry)
			time.Sleep(time.Second * 3)
			retry--
			continue
		}
		break
	}

	var f *os.File
	f, err = os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		log.Println("file open error", err, filePath, retry)
		return nil, err
	}
	return f, nil
}

func getMD5(filePath string) ([]byte, error) {
	fi, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	if fi.IsDir() {
		return nil, errors.New("path is dir")
	}

	return []byte(fi.ModTime().Format(time.RFC3339Nano)), nil

	/*
		f, err := openFile(filePath, time.Hour)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		h := md5.New()
		if _, err := io.Copy(h, f); err != nil {
			return nil, err
		}
		return h.Sum(nil), nil
	*/
}
