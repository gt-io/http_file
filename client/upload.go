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
	"sync"
	"time"
)

var (
	buf  chan string
	done chan bool
)

func startUploader(bufferSize int, wg *sync.WaitGroup) {
	buf = make(chan string, bufferSize)
	go func() {
		defer wg.Done()
		for {
			fn, ok := <-buf
			if !ok {
				log.Println("close uploader")
				return
			}
			log.Println("run...", fn)
			// get md5
			h, err := getMD5(fn)
			if err != nil {
				log.Println("getMD5 error", err, fn)
				continue
			}

			// check aleady uploaded
			if exist, _ := existData(fn, h); exist {
				log.Println("aleady exist data", fn)
				continue
			}

			log.Println("upload start ", fn)

			if err := upload(fn); err != nil {
				log.Println("file upload error", err, fn)
				continue
			}

			addData(fn, h, time.Now())
		}
	}()

}

func closeUploader() {
	close(buf)
}

func post(p string) {
	log.Println("post..", p)

	buf <- p
}

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
			log.Println("openFile error", err, uploadFilePath)
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
	u := conf.URL + "?p=" + url.QueryEscape(p[len(filepath.FromSlash(conf.Path))+1:])

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
