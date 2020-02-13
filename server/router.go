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

	"github.com/google/uuid"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	ip = strings.ReplaceAll(ip, ":", ".")

	paramSavePath := ""
	{
		keys, ok := r.URL.Query()["p"]
		if !ok || len(keys[0]) < 1 {
			log.Println("Url Param 'p' is missing")
		} else {
			p, err := url.QueryUnescape(keys[0])
			if err != nil {
				log.Println("Url parse error", keys[0])
				return
			}
			paramSavePath = p
		}
	}

	saveDir := dstFolder + string(os.PathSeparator) + filepath.FromSlash(strings.ReplaceAll(paramSavePath, "\\", "/")) // time.Now().Format("2006-01-02")

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

		log.Println("new upload file", ip, paramSavePath)
		file, header, err := r.FormFile("myFile")
		if err != nil {
			log.Println("not found form myFile", err)
			fmt.Fprintln(w, err)
			return
		}
		defer file.Close()

		var tempFile string
		if tmpFolder == "" {
			tempFile = filepath.FromSlash(saveDir + string(os.PathSeparator) + header.Filename)
		} else {
			tempFile = tmpFolder + string(os.PathSeparator) + uuid.New().String()
		}

		// make save dir
		if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
			log.Println("create directory error", err, saveDir)
		}

		log.Println("upload start :", header.Filename, tempFile)

		start := time.Now()
		out, err := os.Create(tempFile)
		if err != nil {
			log.Println("Unable to create the file for writing. Check your write access privilege", err, tempFile)
			fmt.Fprintf(w, "Unable to create the file for writing. Check your write access privilege")
			return
		}
		defer func() {
			if out != nil {
				out.Close()
			}
		}()

		// write the content from POST to the file
		_, err = io.Copy(out, file)
		if err != nil {
			log.Println("io.Copy error", err)
			fmt.Fprintln(w, err)
			return
		}

		fmt.Fprintf(w, "File uploaded successfully: ")
		fmt.Fprintf(w, header.Filename)

		if tmpFolder != "" {
			savePath := filepath.FromSlash(saveDir + string(os.PathSeparator) + header.Filename)

			out.Close()
			if err := os.Rename(tempFile, savePath); err != nil {
				log.Println("move file error", err)
				return
			}
			tempFile = savePath
		}

		// save complete log
		l := time.Now().Format(time.RFC3339) + "," + ip + ",\"" + tempFile + "\"\n"
		if _, err := completeFile.WriteString(l); err != nil {
			log.Println("complete log write fail", err)
		}

		log.Println("upload finish :", tempFile, time.Since(start))
	}
}
