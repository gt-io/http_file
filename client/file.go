package main

import (
	"io/ioutil"
	"log"
	"math"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
)

func checkExistFile(watchPath string) error {
	log.Println("start exist file.", watchPath)
	files, err := ioutil.ReadDir(watchPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			checkExistFile(watchPath + "/" + file.Name())
			continue
		}
		post(watchPath + "/" + file.Name())
	}

	log.Println("finish exist file.", watchPath)
	return nil
}

func watchFolder(watchPath string) error {
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
						watchImpl(fn, watcher)
						continue
					}

					post(fn)
				} else if event.Op&fsnotify.Write == fsnotify.Write {
					fn := event.Name

					info, err := os.Stat(fn)
					if os.IsNotExist(err) {
						log.Println("file not exist", fn, os.IsNotExist(err))
						continue
					}
					if info.IsDir() {
						continue
					}

					post(fn)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	watchImpl(watchPath, watcher)

	<-done
	return nil
}

func watchImpl(watchPath string, watcher *fsnotify.Watcher) {
	// main foler watch
	if err := watcher.Add(watchPath); err != nil {
		log.Fatal(err)
	}
	log.Println("watch start folder :", watchPath)

	files, err := ioutil.ReadDir(watchPath)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if file.IsDir() {
			watchImpl(watchPath+"/"+file.Name(), watcher)
		}
	}
}

func openFile(filePath string, wait time.Duration) (*os.File, error) {
	fi, err := os.Stat(filePath)
	if err != nil {
		return nil, err
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
