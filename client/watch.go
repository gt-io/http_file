package main

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/rjeczalik/notify"
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
			checkUploadFiles(watchPath + "/" + file.Name())
		}
	}

	log.Println("finish exist file.", watchPath)
	return nil
}

func watchFolder(watchPath string) {
	c := make(chan notify.EventInfo, 1)
	if err := notify.Watch(watchPath+"/...", c, notify.Create, notify.Write); err != nil {
		log.Fatal(err)
	}
	defer notify.Stop(c)

	timer := time.NewTicker(time.Second * 1)

	done := make(chan bool)
	go func() {
		for {
			select {
			case <-timer.C:
				watchProc()
			case ev, ok := <-c:
				if !ok {
					log.Println("watcher close")
					return
				}
				log.Println("event:", ev.Event(), ev.Path())
				post(ev.Path())
			}
		}
	}()
	<-done
}

var errPath string

func watchProc() {
	// 1. check exist file
	_, err := os.Stat(syncPath)
	if err != nil {
		return
	}

	// 2. read check folder from sync file
	data, err := ioutil.ReadFile(syncPath)
	if err != nil {
		log.Println("read file", err)
		return
	}

	// 3. check dir
	var checkDir string
	if data != nil && len(data) > 0 {
		checkDir = string(data)
		if checkDir != errPath {
			log.Println("new check dir!", checkDir)
			checkUploadFiles(conf.Path + "/" + checkDir)
		}
	} else {
		log.Println("sync file is empty")
		return
	}

	// 4. delete syncfile
	if err := os.Remove(syncPath); err != nil {
		log.Println("sync file remove fail", err, checkDir)
		errPath = checkDir
	} else {
		errPath = ""
	}
}
