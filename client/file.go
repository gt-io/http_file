package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
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
			checkExistFile(path + "/" + file.Name())
			continue
		}
		fn := path + "/" + file.Name()

		// get md5
		h, err := getMD5(fn)
		if err != nil {
			log.Println("getMD5 error", fn)
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
					fn := event.Name

					info, err := os.Stat(fn)
					if os.IsNotExist(err) {
						log.Println("file not exist", fn, os.IsNotExist(err))
						continue
					}
					if info.IsDir() {
						err = watcher.Add(fn)
						if err != nil {
							log.Fatal(err)
						}
						log.Println("watch start folder :", fn)
						continue
					}

					// get md5
					h, err := getMD5(fn)
					if err != nil {
						log.Println("getMD5 error", fn)
						continue
					}

					// check aleady uploaded
					if exist, _ := existData(fn, h); exist {
						log.Println("aleady exist data", fn)
						continue
					}

					log.Println("upload start ", fn)

					if err := upload(fn); err != nil {
						log.Println("file upload error", err)
						continue
					}

					addData(fn, h, time.Now())

				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	watchImpl(path, watcher)

	<-done
	return nil
}

func watchImpl(path string, watcher *fsnotify.Watcher) {
	// main foler watch
	if err := watcher.Add(path); err != nil {
		log.Fatal(err)
	}
	log.Println("watch start folder :", path)

	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if file.IsDir() {
			watchImpl(path+"/"+file.Name(), watcher)
		}
	}
}

func openFile(path string, wait time.Duration) (*os.File, error) {
	var f *os.File

	retry := math.Ceil(wait.Seconds() / 3)

	var err error
	for retry > 0 {
		f, err = os.OpenFile(path, os.O_RDWR, 0644)
		if err != nil {
			log.Println("file open error", err, path, retry)
			time.Sleep(time.Second * 3)
			retry--
			continue
		}

		if fi, _ := f.Stat(); time.Since(fi.ModTime()) < time.Second*10 {
			log.Println("file is busy", path, retry)
			time.Sleep(time.Second * 3)
			retry--
			continue
		}

		break
	}
	if f == nil {
		return nil, fmt.Errorf("file open error %s", path)
	}
	return f, nil
}

func getMD5(path string) ([]byte, error) {
	f, err := openFile(path, time.Hour)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}
