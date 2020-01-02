package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

var dstFolder string

func main() {
	flag.StringVar(&dstFolder, "path", ".", "save folder")
	flag.Parse()
	log.Println("save file path", dstFolder)

	http.HandleFunc("/upload", uploadHandler) // Display a form for user to upload file
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	ip = strings.ReplaceAll(ip, ":", ".")

	log.Println("new upload file", ip)
	file, header, err := r.FormFile("myFile")
	if err != nil {
		log.Println("not found form myFile", err)
		fmt.Fprintln(w, err)
		return
	}
	defer file.Close()

	log.Println("upload start :", header.Filename)

	saveDir := dstFolder + "/" + time.Now().Format("2006-01-02")
	os.MkdirAll(saveDir, os.ModePerm)

	savePath := fmt.Sprintf("%s//%s-%d-%s", saveDir, ip, time.Now().Unix(), header.Filename)
	log.Println("save to", savePath)

	start := time.Now()
	out, err := os.Create(savePath)
	if err != nil {
		log.Println("Unable to create the file for writing. Check your write access privilege", err, savePath)
		fmt.Fprintf(w, "Unable to create the file for writing. Check your write access privilege")
		return
	}
	defer out.Close()

	// write the content from POST to the file
	_, err = io.Copy(out, file)
	if err != nil {
		log.Println("io.Copy error", err)
		fmt.Fprintln(w, err)
		return
	}

	fmt.Fprintf(w, "File uploaded successfully: ")
	fmt.Fprintf(w, header.Filename)

	log.Println("upload finish :", savePath, time.Since(start))

}
