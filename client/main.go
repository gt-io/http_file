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

	"github.com/fsnotify/fsnotify"
)

var url string

func main() {
	// open file db
	ex, _ := os.Executable()
	if err := InitDB(filepath.Dir(ex) + "/data.db"); err != nil {
		panic(err)
	}

	var path string

	flag.StringVar(&url, "url", "http://localhost:8080/upload", "upload path")
	flag.StringVar(&path, "path", "", "watch path")

	flag.Parse()

	checkExistFile(path)
	watch(path)
}

func checkExistFile(path string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// check aleady uploaded
		if exist, _ := ExistData(file.Name()); exist {
			log.Println("aleady exist data", file.Name())
			continue
		}

		log.Println("upload start ", file.Name())

		if err := upload(file.Name()); err != nil {
			log.Println("file upload error", err, file.Name())
			continue
		}

		AddData(file.Name(), "", time.Now())
	}
	return nil
}

func watch(path string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Create == fsnotify.Create {
					log.Println("create file:", event.Name)
					fn := event.Name

					// check file exist
					if info, err := os.Stat(fn); os.IsNotExist(err) || info.IsDir() {
						log.Fatalln("file not exist", fn, os.IsNotExist(err))
						return
					}

					// check aleady uploaded
					if exist, _ := ExistData(fn); exist {
						log.Println("aleady exist data", fn)
						return
					}

					log.Println("upload start ", fn)

					if err := upload(fn); err != nil {
						log.Println("file upload error")
						return
					}

					AddData(fn, "", time.Now())

				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(path)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("watch start folder :", path)
	<-done

	return nil
}

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
