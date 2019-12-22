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
)

func main() {
	var url, path string

	flag.StringVar(&url, "url", "http://localhost:8080/upload", "upload path")
	flag.StringVar(&path, "path", "", "upload file")

	flag.Parse()

	log.Println("upload start ", url, path)

	upload(url, path)
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
		file, err := os.Open(path)
		if err != nil {
			log.Println("upload file os.Open error", err, path)
			return
		}
		defer file.Close()
		if _, err = io.Copy(part, file); err != nil {
			log.Println("io.Copy error", err)
			return
		}
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
