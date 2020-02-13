package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
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
			checkExistFile(watchPath + string(os.PathSeparator) + file.Name())
			checkUploadFiles(watchPath + string(os.PathSeparator) + file.Name())
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

	fo, err := os.Open(syncPath)
	if err != nil {
		log.Println("file open error", err, syncPath)
		return
	}

	reader := bufio.NewReader(fo)
	for {
		line, isPrefix, err := reader.ReadLine()
		if isPrefix || err != nil {
			log.Println("read line error", err, isPrefix)
			break
		}
		checkDir := string(line)
		strings.TrimSuffix(checkDir, "\n")
		strings.TrimSuffix(checkDir, "\r")
		if checkDir == "" {
			break
		}
		log.Println("new check dir!", checkDir)
		checkUploadFiles(conf.Path + string(os.PathSeparator) + filepath.FromSlash(checkDir))
	}
	fo.Close()

	// 4. delete syncfile
	if err := os.Remove(syncPath); err != nil {
		log.Println("sync file remove fail", err)
	} else {
		log.Println("sync file removed")
	}
}
