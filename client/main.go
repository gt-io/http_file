package main

import (
	"flag"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func main() {
	// open file db
	ex, _ := os.Executable()
	if err := InitDB(filepath.Dir(ex) + "/data.db"); err != nil {
		panic(err)
	}

	var url, path string

	flag.StringVar(&url, "url", "http://localhost:8080/upload", "upload path")
	flag.StringVar(&path, "path", "", "upload file")

	flag.Parse()

	// check file exist
	if info, err := os.Stat(path); os.IsNotExist(err) || info.IsDir() {
		log.Fatalln("file not exist", path, os.IsNotExist(err))
		return
	}

	// check aleady uploaded
	if exist, _ := ExistData(path); exist {
		log.Println("aleady exist data", path)
		return
	}

	log.Println("upload start ", url, path)

	if err := upload(url, path); err != nil {
		log.Println("file upload error")
		return
	}

	AddData(path, "", time.Now())

}

func upload(url, path string) error {
	r, w := io.Pipe()
	m := multipart.NewWriter(w)
	go func() {
		defer w.Close()
		defer m.Close()
		part, err := m.CreateFormFile("myFile", filepath.Base(path))
		if err != nil {
			log.Println("multipart CreateFormFile error", err)
			return
		}
		log.Println("file open start",path)
		file, err := os.Open(path)
		if err != nil {
			log.Println("upload file os.Open error", err, path)
			return
		}
		defer file.Close()

		log.Println("file copy start",path)
		if _, err = io.Copy(part, file); err != nil {
			log.Println("io.Copy error", err)
			return
		}
		log.Println("file copy finish",path)
	}()


	resp, err := http.Post(url, m.FormDataContentType(), r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 결과 출력
	bytes, _ := ioutil.ReadAll(resp.Body)
	log.Println("upload ok", string(bytes))
	return nil
}
