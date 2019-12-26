package main

import (
	"flag"
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

	flag.StringVar(&url, "url", "http://localhost:8080/upload", "upload path")
	flag.StringVar(&p.path, "path", "", "watch path")

	flag.Parse()

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
