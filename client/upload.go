package main

import (
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

func upload(uploadFilePath string) error {
	r, w := io.Pipe()
	m := multipart.NewWriter(w)
	go func() {
		defer w.Close()
		defer m.Close()
		part, err := m.CreateFormFile("myFile", filepath.Base(uploadFilePath))
		if err != nil {
			log.Println("multipart CreateFormFile error", err)
			return
		}

		var file *os.File

		log.Println("file open start", uploadFilePath)
		file, err = openFile(uploadFilePath, time.Hour)
		if err != nil {
			return
		}
		defer file.Close()

		log.Println("file copy start", uploadFilePath)
		if _, err = io.Copy(part, file); err != nil {
			log.Println("io.Copy error", err)
			return
		}
		log.Println("file copy finish", uploadFilePath)
	}()

	// parse url.
	p := filepath.Dir(uploadFilePath)
	u := surl + "?p=" + url.QueryEscape(p[len(filepath.FromSlash(wpath))+1:])

	resp, err := http.Post(u, m.FormDataContentType(), r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 결과 출력
	bytes, _ := ioutil.ReadAll(resp.Body)
	log.Println("upload ok", u, uploadFilePath, string(bytes))
	return nil
}
