package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/mailgun/log"
	"io/ioutil"
	"net/http"
	"os"
)

var useFileCache bool
var tmpDir string
var version string

func main() {

	// Set up the logger.
	console, _ := log.NewLogger(log.Config{"console", "info"})
	syslog, _ := log.NewLogger(log.Config{"syslog", "error"})
	log.Init(console, syslog)

	// Set up command line flags.
	addr := flag.String("listen", "0.0.0.0:3000", "IP address/port to listen on")
	cacheDir := flag.String("cache-dir", fmt.Sprintf("%s", os.TempDir()), "Directory for caching static resources")
	cdnPrefix := flag.String("prefix-cdn", "cdn", "The prefix for the fetching CDN resources")
	imgPrefix := flag.String("prefix-img", "img", "The prefix for handling image requests")

	flag.BoolVar(&useFileCache, "cache", true, "Enable filesystem caching")
	flag.Parse()

	tmpDir = *cacheDir
	version = "0.1.0"

	// Make sure cache directory exists.
	// and check if we can write to file cache.
	if useFileCache {
		if err := os.MkdirAll(tmpDir, os.FileMode(0775)); err != nil {
			useFileCache = false
			log.Errorf("Couldn't create tmp dir %s: %s", tmpDir, err)
		} else {
			if err := ioutil.WriteFile(fmt.Sprintf("%s/lock", tmpDir), []byte(version), 0664); err != nil {
				useFileCache = false
				log.Errorf("Couldn't write to tmp dir %s: %s", tmpDir, err)
			}
		}
	}

	log.Infof("Ready to listen on %s", *addr)
	log.Infof("Serving cdn files from %s/%s", *addr, *cdnPrefix)
	log.Infof("Serving external images from %s/%s ", *addr, *imgPrefix)
	if useFileCache {
		log.Infof("Caching via filesystem enabled")
		log.Infof("Using %v as cache path", tmpDir)
	}

	r := mux.NewRouter()
	r.HandleFunc(`/`+*cdnPrefix+`/{package}/{version}/{path:[^?]+}`, cdnHandler)
	r.HandleFunc(`/`+*imgPrefix+`/{domain}/{path:[^?]+}`, respImgHandler)

	log.Errorf("Server died: %s", http.ListenAndServe(*addr, handlers.CompressHandler(r)))

}
