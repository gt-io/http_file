package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
)

var (
	dstFolder    string
	tmpFolder    string
	completeFile *os.File
)

// Config ...
type Config struct {
	Path    string `json:"path"`
	Port    string `json:"port"`
	TempDir string `json:"temp_dir"`
}

func loadConfig(configPath string) (string, string, string, error) {
	var c Config

	if configPath != "" {
		// load from conf.json
		dat, err := ioutil.ReadFile(configPath)
		if err != nil {
			return "", "", "", err
		}
		if err = json.Unmarshal(dat, &c); err != nil {
			return "", "", "", err
		}

		log.Println("start server", c)
		return c.Path, c.Port, c.TempDir, nil
	}

	flag.StringVar(&c.Path, "path", ".", "upload path")
	flag.StringVar(&c.Port, "port", ":8081", "listen addr:port")
	flag.StringVar(&c.TempDir, "temp_dir", "", "temporary directory")

	flag.Parse()
	log.Println("start server", c)
	return c.Path, c.Port, c.TempDir, nil
}
