/*
Copyright 2022 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package routers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
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
	WebUIPath     string `json:"WebUIPath"`
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
	return writeConfig(r.configFile, r.Config)
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

func loadPort() (string, error) {
	if len(os.Args) < 2 {
		return "", fmt.Errorf("port not provided")
	}
	port := os.Args[1]
	fmt.Println("Port:", port)
	return port, nil
}

func loadConfig(filePath string) (Config, error) {
	var config Config
	data, err := os.ReadFile(filePath)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}

func (r *Routers) ListenConfig() {
	go file(func(e fsnotify.Event) {
		if e.Has(fsnotify.Write) && e.Name == r.configFile {
			config, err := loadConfig(r.configFile)
			if err != nil {
				fmt.Println("Error reading config file:", err)
				return
			}
			r.Config = config
			fmt.Println("Config file updated")
		}
	}, r.configFile)
}

// Watch one or more files, but instead of watching the file directly it watches
// the parent directory. This solves various issues where files are frequently
// renamed, such as editors saving them.
func file(callback func(fsnotify.Event), files ...string) {
	if len(files) < 1 {
		panic("must specify at least one file to watch")
	}

	// Create a new watcher.
	w, err := fsnotify.NewWatcher()
	if err != nil {
		panic(fmt.Errorf("creating a new watcher: %s", err))
	}
	defer w.Close()

	// Start listening for events.
	go fileLoop(w, files, callback)

	// Add all files from the commandline.
	for _, p := range files {
		st, err := os.Lstat(p)
		if err != nil {
			panic(fmt.Errorf("%s", err))
		}

		if st.IsDir() {
			panic(fmt.Errorf("%q is a directory, not a file", p))
		}

		// Watch the directory, not the file itself.
		err = w.Add(filepath.Dir(p))
		if err != nil {
			panic(fmt.Errorf("%q: %s", p, err))
		}
	}

	// Block forever
	<-make(chan struct{})
}

func fileLoop(w *fsnotify.Watcher, files []string, callback func(fsnotify.Event)) {

	var debounceTimer *time.Timer
	debounceDuration := 2 * time.Second // 防抖时间，可以根据需要调整

	processEvent := func(e fsnotify.Event) {
		for _, f := range files {
			if filepath.Base(f) == filepath.Base(e.Name) {
				callback(e)
				break
			}
		}
	}

	for {
		select {
		// Read from Errors.
		case err, ok := <-w.Errors:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}
			fmt.Printf("ERROR: %s", err)
		// Read from Events.
		case e, ok := <-w.Events:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}

			// 如果定时器已经存在，停止并重置定时器
			if debounceTimer != nil {
				debounceTimer.Stop()
			}

			// 创建或重置定时器，当定时器触发时调用processEvent处理事件
			debounceTimer = time.AfterFunc(debounceDuration, func() {
				processEvent(e)
			})
		}
	}
}
