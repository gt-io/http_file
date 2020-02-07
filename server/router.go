package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	ip = strings.ReplaceAll(ip, ":", ".")

	keys, ok := r.URL.Query()["p"]
	if !ok || len(keys[0]) < 1 {
		log.Println("Url Param 'p' is missing")
		return
	}
	p, err := url.QueryUnescape(keys[0])
	if err != nil {
		log.Println("Url parse error", keys[0])
		return
	}
	saveDir := dstFolder + "/" + strings.ReplaceAll(filepath.ToSlash(p), "\\", "/") // time.Now().Format("2006-01-02")

	switch r.Method {
	case "GET":
		// check exist path.
		if _, err := os.Stat(saveDir); os.IsNotExist(err) {
			log.Println("dir is not exist", saveDir)
			// path/to/whatever does not exist
			// send response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]int64{})
			return
		}

		// get file list.
		files, err := ioutil.ReadDir(saveDir)
		if err != nil {
			log.Println("read dir fail", err, saveDir)
			return
		}

		result := make(map[string]int64)
		for _, f := range files {
			if !f.IsDir() {
				result[f.Name()] = f.Size()
			}
		}

		// send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(result)

	case "POST":

		log.Println("new upload file", ip, p)
		file, header, err := r.FormFile("myFile")
		if err != nil {
			log.Println("not found form myFile", err)
			fmt.Fprintln(w, err)
			return
		}
		defer file.Close()

		log.Println("upload start :", dstFolder, header.Filename)

		if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
			log.Println("create drirect error", err, saveDir)
		}

		savePath := filepath.FromSlash(fmt.Sprintf("%s/%s", saveDir, header.Filename))
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

		l := time.Now().Format(time.RFC3339) + "," + ip + ",\"" + savePath + "\"\n"
		if _, err := completeFile.WriteString(l); err != nil {
			log.Println(err)
		}

		log.Println("upload finish :", savePath, time.Since(start))
	}
}
