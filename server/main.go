package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

	log.Println("new upload file")
	file, header, err := r.FormFile("myFile")
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	defer file.Close()

	log.Println("upload start :", header.Filename)

	saveDir := dstFolder + "/" + time.Now().Format("2006-01-02")
	os.MkdirAll(saveDir, os.ModePerm)

	savePath := saveDir + "/" + header.Filename

	start := time.Now()
	out, err := os.Create(savePath)
	if err != nil {
		fmt.Fprintf(w, "Unable to create the file for writing. Check your write access privilege")
		return
	}
	defer out.Close()

	// write the content from POST to the file
	_, err = io.Copy(out, file)
	if err != nil {
		fmt.Fprintln(w, err)
	}

	fmt.Fprintf(w, "File uploaded successfully: ")
	fmt.Fprintf(w, header.Filename)

	log.Println("upload finish :", savePath, time.Since(start))

}
