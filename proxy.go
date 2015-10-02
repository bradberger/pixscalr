package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mailgun/log"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"path"
	"strings"
)

func proxyHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	domain := vars["domain"]
	reqPath := vars["path"]
	if domain == "github.com" {
		domain = "raw.githubusercontent.com"
		reqPath = strings.Replace(reqPath, "/blob", "", 1)
	}

	url := fmt.Sprintf("%s/%s", domain, reqPath)
	cachePath := fmt.Sprintf("%s/%s", tmpDir, url)

	// Check for file in cache.
	if useFileCache {
		contents, err := ioutil.ReadFile(path.Clean(cachePath))
		if err == nil {

			log.Infof("HIT %s", cachePath)
			addCacheHeader(&w, "yes")
			w.Header().Set("Content-Type", mime.TypeByExtension(path.Ext(reqPath)))
			w.Write(contents)

			return
		}
	}

	resp, err := http.Get(fmt.Sprintf("http://%s", url))
	if err != nil {
		log.Infof("ERROR %s", err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}

	if resp.StatusCode >= 400 {
		http.Error(w, resp.Status, resp.StatusCode)
		return
	}

	log.Infof("PROXY %s", url)
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))

	if useFileCache {
		writeAndCache(w, resp.Body, cachePath)
		return
	}

	io.Copy(w, resp.Body)
	return

}
