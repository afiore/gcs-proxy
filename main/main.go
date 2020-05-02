package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/afiore/gcs-proxy/config"
	"github.com/afiore/gcs-proxy/gcs"
	"github.com/afiore/gcs-proxy/server"
	"github.com/influxdata/toml"
)

func main() {
	progName := os.Args[0]
	if len(os.Args) < 2 {
		os.Stderr.WriteString(fmt.Sprintf("usage: %s </path/to/config.toml>", progName))
		os.Exit(1)
		return
	}
	configPath := os.Args[1]
	configBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("usage: %s </path/to/config.toml>", progName)
	}

	var config config.ProgramConfig
	err = toml.Unmarshal(configBytes, &config)
	if err != nil {
		log.Fatalf("Couldn't parse configuration file: %s", err)
	}

	store := gcs.StoreOps(config.Gcs.ServiceAccountFilePath)
	http.HandleFunc("/", server.ServeFromBuckets(config.Gcs.Buckets, store))
	log.Printf("Loading server with config: %-v", config)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Web.Port), nil))
}
