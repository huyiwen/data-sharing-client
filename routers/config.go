/*
Copyright 2022 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package routers

import (
	"encoding/json"
	"fmt"
	"os"
)

type ServiceCredentials struct {
	DatabaseIP       string `json:"DatabaseIP"`
	DatabasePort     string `json:"DatabasePort"`
	DatabaseUser     string `json:"DatabaseUser"`
	DatabasePassword string `json:"DatabasePassword"`
	DatabaseName     string `json:"DatabaseName"`
	DatabaseTable    string `json:"DatabaseTable"`
}

type ServiceInformation struct {
	DisplayName string `json:"DisplayName"`
	Description string `json:"Description"`
}

type ServiceType struct {
	Information ServiceInformation `json:"Information"`
	Credentials ServiceCredentials `json:"Credentials"`
}

type Config struct {
	QueryContract struct {
		ChaincodeName string `json:"ChaincodeName"`
		ChannelID     string `json:"ChannelID"`
	} `json:"QueryContract"`
	ServiceContract struct {
		ChaincodeName string `json:"ChaincodeName"`
		ChannelID     string `json:"ChannelID"`
	} `json:"ServiceContract"`
	Services map[string]ServiceType `json:"Services"`
}

func (r *Routers) updateConfig() error {
	return writeConfig(r.configPath, r.Config)
}

func writeConfig(path string, config Config) error {
	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func loadConfig(path string) (Config, string, error) {
	var config Config
	data, err := os.ReadFile(path)
	if err != nil {
		return config, "", err
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, "", err
	}
	if len(os.Args) < 2 {
		return config, "", fmt.Errorf("port not provided")
	}
	port := os.Args[1]
	fmt.Println("Port:", port)
	return config, port, nil
}
