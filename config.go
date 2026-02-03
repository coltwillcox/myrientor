package main

import (
	"encoding/json"
	"os"
)

type Device struct {
	RemotePath string `json:"remote_path"`
	Sync       bool   `json:"sync"`
	LocalPath  string `json:"local_path"`
}

type Config struct {
	BaseURL string   `json:"base_url"`
	Devices []Device `json:"devices"`
}

func readConfigFile(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
