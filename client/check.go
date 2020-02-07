package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
)

func getServerFileInfo(p string) (map[string]int64, error) {
	u := conf.URL + "?p=" + url.QueryEscape(p[len(filepath.FromSlash(conf.Path))+1:])

	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]int64
	err = json.NewDecoder(resp.Body).Decode(&result)

	return result, nil
}

func getLocalFileInfo(watchPath string) (map[string]int64, error) {
	// get file list.
	files, err := ioutil.ReadDir(watchPath)
	if err != nil {
		log.Println("read dir fail", err, watchPath)
		return nil, err
	}

	result := make(map[string]int64)
	for _, f := range files {
		if !f.IsDir() {
			result[f.Name()] = f.Size()
		}
	}
	return result, nil
}

func checkUploadFiles(p string) error {
	// get local file info.
	lfiles, err := getLocalFileInfo(conf.Path + "/" + p)
	if err != nil {
		return err
	}

	// get server file info.
	sfiles, err := getServerFileInfo(conf.Path + "/" + p)
	if err != nil {
		return err
	}

	// compare files
	isSame := true
	for lfn, lsize := range lfiles {
		ssize, ok := sfiles[lfn]
		if ok && lsize == ssize {
			log.Println("same file!", lfn, lsize)
			continue
		}
		isSame = false
		// diff..upload
		log.Println("diff file", lfn, lsize, ssize)

		delData(conf.Path + "/" + p + "/" + lfn)

		post(conf.Path + "/" + p + "/" + lfn)
	}

	if isSame {
		log.Println("same directory", p)
	}

	return nil
}
