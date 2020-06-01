package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/afiore/gcs-proxy/config"
	"github.com/afiore/gcs-proxy/gcs"
	"github.com/afiore/gcs-proxy/server"
)

func main() {
	var conf config.ProgramConfig

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

	err = toml.Unmarshal(configBytes, &conf)
	if err != nil {
		log.Fatalf("Couldn't parse configuration file: %s", err)
	}

	store := gcs.StoreOps(conf.Gcs.ServiceAccountFilePath)
	gcsHandler := server.ServeFromBuckets(conf.Gcs.Buckets, store)
	serverHandler := server.ValidatingSession(conf.Web.OAuth.AllowedHostDomains, conf.Web.OAuth.SessionSecret, gcsHandler)

	mux := http.NewServeMux()

	googleOAuth := server.Handlers(conf)

	mux.HandleFunc("/", serverHandler)
	mux.HandleFunc("/auth/google/login", googleOAuth.Login)
	mux.HandleFunc("/auth/google/callback", googleOAuth.Callback)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", conf.Web.Port),
		Handler: mux,
	}

	http.HandleFunc("/", serverHandler)
	log.Printf("Loading server with config: %+v", conf)
	log.Fatal(httpServer.ListenAndServe())
}
