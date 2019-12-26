package main

import (
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func upload(path string) error {
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
		log.Println("file open start", path)
		file, err := os.Open(path)
		if err != nil {
			log.Println("upload file os.Open error", err, path)
			return
		}
		defer file.Close()

		log.Println("file copy start", path)
		if _, err = io.Copy(part, file); err != nil {
			log.Println("io.Copy error", err)
			return
		}
		log.Println("file copy finish", path)
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
