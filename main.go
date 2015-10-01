package main

import (
    "net/http"
    "github.com/gorilla/mux"
    "github.com/mailgun/log"
    "flag"
    "os"
    "fmt"
    "io/ioutil"
    "github.com/gorilla/handlers"
)

var useFileCache bool
var tmpDir string
var version string

func main() {

    // Set up command line flags.
    addr := flag.String("listen", ":3000", "IP address/port to listen on")
    cacheDir := flag.String("cache", fmt.Sprintf("%s/gofaster", os.TempDir()), "Directory for caching static resources")
    cdnPrefix := flag.String("cdnprefix", "cdn", "The prefix for the fetching CDN resources")
    imgPrefix := flag.String("imgprefix", "img", "The prefix for handling image requests")

    flag.Parse()

    tmpDir = *cacheDir
    version = "0.1.0"

    // Make sure cache directory exists.
    err := os.MkdirAll(tmpDir, os.FileMode(0775)); if err != nil {
        tmpDir = ""
        log.Errorf("Couldn't create tmp dir %s: %s", tmpDir, err)
    }

    // Now check if we can write to file cache.
    err = ioutil.WriteFile(fmt.Sprintf("%s/lock", tmpDir), []byte(version), 0664)
    if err != nil {
        tmpDir = ""
        log.Errorf("Couldn't write to tmp dir %s: %s", tmpDir, err)
    }

    // If no tmpDir, then we can't write tmp files, so note that.
    useFileCache = tmpDir != ""

    // Set up the logger.
    console, _ := log.NewLogger(log.Config{"console", "info"})
    syslog, _ := log.NewLogger(log.Config{"syslog", "error"})
    log.Init(console, syslog)

    r := mux.NewRouter()
    r.HandleFunc(`/` + *cdnPrefix + `/{package}/{version}/{path:[^?]+}`, cdnHandler)
    r.HandleFunc(`/` + *imgPrefix + `/{domain}/{path:[^?]+}`, respImgHandler)

    log.Infof("Using %v as cache path", tmpDir)
    log.Infof("Listening on %s", *addr)
    log.Errorf("Server died: %s", http.ListenAndServe(*addr, handlers.CompressHandler(r)))

}
