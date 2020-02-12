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

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]int64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

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

func checkUploadFiles(p string) (bool, error) {
	// get local file info.
	lfiles, err := getLocalFileInfo(p)
	if err != nil {
		return false, err
	}

	if lfiles == nil || len(lfiles) == 0 {
		return true, nil
	}

	// get server file info.
	sfiles, err := getServerFileInfo(p)
	if err != nil {
		return false, err
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

		delData(p + "/" + lfn)

		post(p + "/" + lfn)
	}

	if isSame {
		log.Println("same directory", p)
	}

	return isSame, nil
}
