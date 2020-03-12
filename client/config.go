package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
)

// Config ...
type Config struct {
	URL   string `json:"url"`
	Path  string `json:"path"`
	Check string `json:"check"`
	Day   int    `json:"day"`
}

func (c *Config) vaild() bool {
	return c.URL != "" && c.Path != ""
}

func loadConfig(configPath string) (*Config, error) {
	var c Config

	flag.StringVar(&c.URL, "url", "http://localhost:8081/upload", "upload path")
	flag.StringVar(&c.Path, "path", "", "watch path")
	flag.StringVar(&c.Check, "check", "", "check path in file server")
	flag.IntVar(&c.Day, "day", 0, "check day")

	flag.Parse()

	if !c.vaild() {
		// load from conf.json
		dat, err := ioutil.ReadFile(configPath)
		if err != nil {
			return nil, err
		}
		if err = json.Unmarshal(dat, &c); err != nil {
			return nil, err
		}

		if !c.vaild() {
			return nil, errors.New("invalid config")
		}
	}

	return &c, nil
}
