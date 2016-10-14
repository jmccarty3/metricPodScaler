package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	api "github.com/jmccarty3/metricPodScaler/api"
	_ "github.com/jmccarty3/metricPodScaler/api/providers/all"
	"gopkg.in/yaml.v2"
)

/*
Load config file
Setup Scalers
Connect to Providers(S)

Run each scaler in go routine
*/

var (
	argConfigFile = flag.String("config", "", "Path to the configuration file")
)

func main() {
	flag.Parse()

	if *argConfigFile == "" {
		panic("No config file given")
	}

	var config api.Config

	configData, err := ioutil.ReadFile(*argConfigFile)
	if err != nil {
		panic(fmt.Sprintf("Error loading conifg file: %v", err))
	}

	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		panic(fmt.Sprintf("Error parsing config file. %v", err))
	}

	for _, scaler := range config.Scalers {
		scaler.Init(config.MasterURL, config.KubeConfig)
		if err := scaler.Provider.Connect(); err != nil {
			panic(fmt.Sprintf("Scaler: %s Error connecting to Provider: %v", scaler.Object.Name, err))
		}
		go scaler.Run()
	}

	select {}
}
