package main

import (
	"crypto/tls"
	"fmt"
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

	httpClient *http.Client
)

func init() {
	var defaultTransport http.Transport

	// Customize the Transport to have larger connection pool
	defaultRoundTripper := http.DefaultTransport
	defaultTransportPointer, ok := defaultRoundTripper.(*http.Transport)
	if !ok {
		panic(fmt.Sprintf("defaultRoundTripper not an *http.Transport"))
	}
	defaultTransport = *defaultTransportPointer // dereference it to get a copy of the struct that the pointer points to
	defaultTransport.MaxIdleConns = 1000
	defaultTransport.MaxIdleConnsPerHost = 1000
	defaultTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	httpClient = &http.Client{
		Transport: &defaultTransport,
		Timeout:   0,
	}
}

func startUploader(bufferSize int, wg *sync.WaitGroup) {
	buf = make(chan string, bufferSize)
	var f *os.File
	go func() {
		defer wg.Done()
		for {
			fn, ok := <-buf
			if !ok {
				log.Println("close uploader")
				return
			}

			//
			if f == nil {
				stateF, err := os.OpenFile(statePath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					log.Println(err)
					continue
				}
				f = stateF
			}
			f.WriteString(fn + "\n")

			uploadProc(fn)

			if len(buf) == 0 {
				f.Close()
				os.Remove(statePath)
				f = nil
			}
		}
	}()
}

func closeUploader() {
	log.Println(" close queue..")
	close(buf)
}

func post(p string) {
	log.Println("post..", p)

	buf <- p
}

func uploadProc(fn string) error {
	log.Println("run....", fn)

	/*
		///////////////////////////////////////////////
		// create state file.
		f, err := os.OpenFile(statePath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
			return err
		}
		defer func() {
			f.Close()
			os.Remove(statePath)
		}()
		f.WriteString(fn)
		///////////////////////////////////////////////
	*/

	// get md5
	h, err := getMD5(fn)
	if err != nil {
		log.Println("getMD5 error", err, fn)
		return err
	}

	// check aleady uploaded
	if exist, _ := existData(fn, h); exist {
		log.Println("aleady exist data", fn)
		return err
	}

	// start upload file
	if err := upload(fn); err != nil {
		log.Println("file upload error", err, fn)
		return err
	}

	// add db
	addData(fn, h, time.Now())

	return nil
}

func upload(uploadFilePath string) error {
	log.Println("upload..", uploadFilePath)

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

		//log.Println("file open start", uploadFilePath)
		file, err = openFile(uploadFilePath, time.Hour)
		if err != nil {
			log.Println("openFile error", err, uploadFilePath)
			return
		}
		defer file.Close()

		//log.Println("file copy start", uploadFilePath)
		if _, err = io.Copy(part, file); err != nil {
			log.Println("io.Copy error", err)
			return
		}
		//log.Println("file copy finish", uploadFilePath)
	}()

	// get save dir.
	saveDir := ""
	p := filepath.Dir(uploadFilePath)
	if p != filepath.FromSlash(conf.Path) {
		saveDir = p[len(filepath.FromSlash(conf.Path))+1:]
	}

	u := conf.URL + "?p=" + url.QueryEscape(saveDir)

	req, err := http.NewRequest("POST", u, r)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", m.FormDataContentType())

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 결과 출력
	bytes, _ := ioutil.ReadAll(resp.Body)
	log.Println("upload ok", uploadFilePath, u, string(bytes))
	return nil
}
