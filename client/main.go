package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/judwhite/go-svc/svc"
)

var url string

// program implements svc.Service
type program struct {
	wg   sync.WaitGroup
	quit chan struct{}

	path string
}

func main() {
	// init log
	ex, _ := os.Executable()
	fpLog, err := os.OpenFile(filepath.Dir(ex)+"/watcher.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer fpLog.Close()

	// 파일과 화면에 같이 출력하기 위해 MultiWriter 생성
	log.SetOutput(io.MultiWriter(fpLog, os.Stdout))

	prg := &program{}

	// Call svc.Run to start your program/service.
	if err := svc.Run(prg); err != nil {
		log.Fatal(err)
	}
}

func (p *program) Init(env svc.Environment) error {
	log.Printf("is win service? %v\n", env.IsWindowsService())

	// open file db
	ex, _ := os.Executable()
	if err := initDB(filepath.Dir(ex) + "/data.db"); err != nil {
		panic(err)
	}

	var err error
	url, p.path, err = loadConfig(filepath.Dir(ex) + "/conf.json")
	if err != nil {
		panic(err)
	}

	log.Println("start..", url, p.path)

	return nil
}

func (p *program) Start() error {
	log.Println("Starting...")

	go checkExistFile(p.path)

	go watchFolder(p.path)

	return nil
}

func (p *program) Stop() error {
	log.Printf("Stopped.\n")
	return nil
}
