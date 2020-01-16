package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/judwhite/go-svc/svc"
)

var (
	surl  string
	wpath string
)

// program implements svc.Service
type program struct {
	wg   sync.WaitGroup
	quit chan struct{}

	port string
}

func main() {
	// init log
	ex, _ := os.Executable()
	fpLog, err := os.OpenFile(filepath.Dir(ex)+"/server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer fpLog.Close()

	log.SetOutput(io.MultiWriter(fpLog, os.Stdout))

	prg := &program{}

	// Call svc.Run to start your program/service.
	if err := svc.Run(prg); err != nil {
		log.Fatal(err)
	}
}

func (p *program) Init(env svc.Environment) error {
	log.Printf("is win service? %v\n", env.IsWindowsService())

	// load config
	ex, _ := os.Executable()
	wpath, lport, err := loadConfig(filepath.Dir(ex) + "/conf.json")
	if err != nil {
		log.Fatal(err)
	}
	dstFolder = wpath
	p.port = lport

	return nil
}

func (p *program) Start() error {
	log.Println("Starting...")

	// init complete file.
	var err error
	completeFile, err = os.OpenFile(dstFolder+"/complete.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln(err)
	}

	// route
	http.Handle("/", http.FileServer(http.Dir(dstFolder)))
	http.HandleFunc("/upload", uploadHandler) // Display a form for user to upload file

	// start server
	go http.ListenAndServe(p.port, nil)
	return nil
}

func (p *program) Stop() error {
	log.Printf("Stopped.\n")
	if completeFile != nil {
		completeFile.Close()
	}
	return nil
}
