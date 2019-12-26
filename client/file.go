package main

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
)

func checkExistFile(path string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fn := path + "/" + file.Name()

		// check aleady uploaded
		if exist, _ := existData(fn); exist {
			log.Println("aleady exist data", fn)
			continue
		}

		log.Println("upload start ", fn)

		if err := upload(fn); err != nil {
			log.Println("file upload error", err, fn)
			continue
		}

		addData(fn, "", time.Now())
	}
	return nil
}

func watchFolder(path string) error {
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
					if exist, _ := existData(fn); exist {
						log.Println("aleady exist data", fn)
						return
					}

					log.Println("upload start ", fn)

					if err := upload(fn); err != nil {
						log.Println("file upload error")
						return
					}

					addData(fn, "", time.Now())

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
