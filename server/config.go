package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
)

var dstFolder string
var completeFile *os.File

// Config ...
type Config struct {
	Path string `json:"path"`
	Port string `json:"port"`
}

func loadConfig(configPath string) (string, string, error) {
	var c Config

	if configPath != "" {
		// load from conf.json
		dat, err := ioutil.ReadFile(configPath)
		if err != nil {
			return "", "", err
		}
		if err = json.Unmarshal(dat, &c); err != nil {
			return "", "", err
		}

		log.Println("start server", c.Path, c.Port)
		return c.Path, c.Port, nil
	}

	flag.StringVar(&c.Path, "path", ".", "upload path")
	flag.StringVar(&c.Port, "port", ":8081", "listen addr:port")

	flag.Parse()
	log.Println("start server", c.Path, c.Port)
	return c.Path, c.Port, nil
}
