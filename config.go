package main

import (
	"encoding/json"
	"os"
)

const (
	remoteConfigFile = "remote.json"
	localConfigFile  = "local.json"
)

type LocalConfig struct {
	MaxConcurrent int `json:"max_concurrent"`
}

type RemoteConfig struct {
	BaseURL string   `json:"base_url"`
	Devices []Device `json:"devices"`
}

type Device struct {
	RemotePath string `json:"remote_path"`
	Sync       bool   `json:"sync"`
	LocalPath  string `json:"local_path"`
}

func readLocalConfigFile() (*LocalConfig, error) {
	file, err := os.Open(localConfigFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config LocalConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func readRemoteConfigFile() (*RemoteConfig, error) {
	file, err := os.Open(remoteConfigFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config RemoteConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
