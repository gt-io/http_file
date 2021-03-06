package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	// init log
	fpLog, err := os.OpenFile("/var/log/server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer fpLog.Close()
	log.SetOutput(io.MultiWriter(fpLog, os.Stdout))

	// load config
	ex, _ := os.Executable()
	wpath, lport, tmpDir, err := loadConfig(ex + ".json")
	if err != nil {
		log.Fatal(err)
	}
	dstFolder = wpath

	if tmpDir != "" {
		if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
			log.Println("create temp directory error", err, tmpDir)
		}
		tmpFolder = tmpDir
	}

	// init complete file.
	f, err := os.OpenFile(dstFolder+"/complete.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	completeFile = f

	// route
	http.Handle("/", http.FileServer(http.Dir(dstFolder)))
	http.HandleFunc("/upload", uploadHandler)

	// start server
	log.Fatal(http.ListenAndServe(lport, nil))
}
