/*
This file is part of the simple-spa-server software package.

Copyright (C) 2020 Paul Scherrer Institute, Switzerland

simple-spa-server is licensed under the terms of the GPL v3
or any later version. See LICENSE.md for details.
*/
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type serverConfig struct {
	docRoot       string
	indexDocument string
	bindAddr      string
	useTLS        bool
	certPath      string
	keyPath       string
}

var config serverConfig
var cachedIndexDocument string

// fileExists checks if a path points to a regular file
// i.e. not a directory
func fileExists(fpath string) bool {
	info, err := os.Stat(fpath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// handler will serve a file, if the path matches exactly, or the
// cached index document file otherwise.
func handler(w http.ResponseWriter, r *http.Request) {
	fpath := filepath.Join(config.docRoot, r.URL.Path)
	if fileExists(fpath) {
		http.ServeFile(w, r, fpath)
		return
	}
	fmt.Fprint(w, cachedIndexDocument)
}

func envOrDefault(envKey string, defaultValue string) string {
	if val, present := os.LookupEnv(envKey); present {
		return val
	}
	return defaultValue
}

func initFromEnv() serverConfig {
	useTLS, err := strconv.ParseBool(envOrDefault("SSS_USE_TLS", "0"))
	if err != nil {
		log.Fatal(err)
	}
	docRoot := envOrDefault("SSS_DOCROOT", "/data/docroot")
	cfg := serverConfig{
		docRoot:       docRoot,
		bindAddr:      envOrDefault("SSS_BINDADDR", ":8080"),
		indexDocument: envOrDefault("SSS_INDEXDOC", filepath.Join(docRoot, "index.html")),
		useTLS:        useTLS,
		certPath:      envOrDefault("SSS_TLS_CERTPATH", "/data/conf/cert.pem"),
		keyPath:       envOrDefault("SSS_TLS_KEYPATH", "/data/conf/key.pem"),
	}
	return cfg
}

// cacheIndexDocument will load the index document, expand the process
// environment and cache the result
func cacheIndexDocument(fpath string) {
	contents, err := ioutil.ReadFile(fpath)
	if err != nil {
		log.Fatal(err)
	}
	cachedIndexDocument = os.ExpandEnv(string(contents))
}

func main() {
	appName := envOrDefault("SSS_APP_NAME", "simple-spa-server")
	log.Printf("%s starting up ...\n", appName)
	config = initFromEnv()
	log.Printf("server configuration:\n%#v\n", config)
	cacheIndexDocument(config.indexDocument)
	http.HandleFunc("/", handler)
	log.Println("Listening on: ", config.bindAddr)
	var err error
	if config.useTLS {
		err = http.ListenAndServeTLS(config.bindAddr, config.certPath, config.keyPath, nil)
	} else {
		err = http.ListenAndServe(config.bindAddr, nil)
	}
	log.Fatal(err)
}
