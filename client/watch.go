package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/rjeczalik/notify"
)

var re = regexp.MustCompile("[0-9]{4}-(0[1-9]|1[0-2])-(0[1-9]|[1-2][0-9]|3[0-1])")

// 이 폴더 아래 디렉토리를 다 읽어서.
// 디렉토리 형식이 YYYY-MM-DD 폴더라면.
// 큰순서대로 정렬 후 checkDay 만큼만 확인하자.
func checkExistFile(watchPath string, checkDay int) error {
	log.Println("start exist dir.", watchPath, checkDay)
	files, err := ioutil.ReadDir(watchPath)
	if err != nil {
		log.Fatal(err)
	}

	// 만약 하위폴더가 날짜폴더라면..목록을 얻어오자.
	var dirList []string
	for _, file := range files {
		if file.IsDir() && re.MatchString(file.Name()) {
			dirList = append(dirList, file.Name())
		}
	}

	if len(dirList) > 0 {
		// 날짜폴더이므로 최근 폴더부터 checkDay 만큼 체크.
		sort.Slice(dirList, func(i, j int) bool {
			return dirList[i] > dirList[j]
		})

		log.Println("check dir", dirList)

		cd := checkDay
		for _, d := range dirList {
			cd--

			checkExistFile(watchPath+string(os.PathSeparator)+d, checkDay)
			checkUploadFiles(watchPath + string(os.PathSeparator) + d)

			if cd == 0 {
				break
			}
		}

	} else {
		// 날짜폴더가 아니므로.. 예전처럼 그냥 디렉토리 체크.
		for _, file := range files {
			if file.IsDir() {
				checkExistFile(watchPath+string(os.PathSeparator)+file.Name(), checkDay)
				checkUploadFiles(watchPath + string(os.PathSeparator) + file.Name())
			}
		}
	}

	log.Println("finish exist dir.", watchPath)
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
