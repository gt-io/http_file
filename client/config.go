package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
)

// Config ...
type Config struct {
	URL  string `json:"url"`
	Path string `json:"path"`
}

func loadConfig(path string) (string, string, error) {
	var u, p string

	flag.StringVar(&u, "url", "http://localhost:8080/upload", "upload path")
	flag.StringVar(&p, "path", "", "watch path")

	flag.Parse()

	if u != "" && p != "" {
		return u, p, nil
	}

	// load from conf.json
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return "", "", err
	}

	var c Config
	if err = json.Unmarshal(dat, &c); err != nil {
		return "", "", err
	}

	return c.URL, c.Path, nil
}
