package main

import (
	"bytes"
	"fmt"
	"github.com/bradberger/optimizer"
	"github.com/gorilla/mux"
	"github.com/mailgun/log"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
	_ "golang.org/x/image/webp/nycbcra"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"mime"
	"net/http"
	"path"
)

func cdnHeaders(w *http.ResponseWriter, r *http.Request) {
	if origin := r.Header.Get("Origin"); origin != "" {
		(*w).Header().Set("Access-Control-Allow-Origin", origin)
		(*w).Header().Set("Timing-Allow-Origin", origin)
		(*w).Header().Set("Access-Control-Allow-Methods", r.Method)
		(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
	}
}

func imgHeaders(w *http.ResponseWriter, r *http.Request) {
	cdnHeaders(w, r)
	(*w).Header().Set("Accept-CH", "DPR, Width, Viewport-Width, Downlink")
	(*w).Header().Set("Vary", "Accept, DPR, Width, Save-Data, Downlink")
}

func serveCachedRespImg(w http.ResponseWriter, r *http.Request, domain string, imagePath string, opts optimizer.Options) (err error) {

	cachePath := getOptimizedImgCachePath(domain, imagePath, opts)
	contents, err := ioutil.ReadFile(path.Clean(cachePath))
	if err == nil {

		// Set Content-Type and Cache-Control headers.
		w.Header().Set("Content-Type", opts.Mime)

		imgHeaders(&w, r)
		addCacheHeader(&w, "optimized")

		w.Write(contents)
		return

	}

	return

}

func getFileFromUpstream(fileURL string, useSSL bool) (body []byte, statusCode int, err error) {

	// Try to fetch the image. If not, fail.
	url := fileURL
	if useSSL {
		url = fmt.Sprintf("https://%s", url)
	} else {
		url = fmt.Sprintf("http://%s", url)
	}

	resp, e := http.Get(url)
	if e != nil {
		err = e
		log.Errorf("Couldn't get %s", url)
		statusCode = 500
		return
	}
	defer resp.Body.Close()

	// Make sure the request succeeded
	if resp.StatusCode > 302 {
		log.Errorf("Couldn't fetch %s", url)
		statusCode = resp.StatusCode
		return
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Couldn't read body %s", url)
		statusCode = 500
	}

	return

}

func getImageFromUpstream(domain string, imagePath string) (body []byte, statusCode int, err error) {
	// Check for resized original on disk.
	body, err = getImageFromCache(fmt.Sprintf("%s/%s/%s", tmpDir, domain, imagePath))
	if err != nil {
		body, statusCode, err = getFileFromUpstream(fmt.Sprintf("%s/%s", domain, imagePath), false)
	}
	return
}

func respLocalImgHandler(w http.ResponseWriter, r *http.Request) {

	domain := r.Host
	imagePath := r.RequestURI
	opts := optimizer.Options{
		Mime: mime.TypeByExtension(path.Ext(imagePath)),
	}

	opts.SetFromRequest(r)
	opts.Optimize()

	// Try to fetch cached iamge.
	if err := serveCachedRespImg(w, r, domain, imagePath, opts); err == nil {
		imgHeaders(&w, r)
		return
	}

	body, err := getImageFromCache(fmt.Sprintf("%s/%s", documentRoot, imagePath))
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}

	optimizeAndServeImg(w, r, opts, domain, imagePath, body)

}

func optimizeAndServeImg(w http.ResponseWriter, r *http.Request, opts optimizer.Options, domain string, imagePath string, body []byte) (err error) {

	img, _, err := image.Decode(bytes.NewReader(body))
	if err != nil || img == nil {
		log.Errorf("Could not decode %s/%s: %s", domain, imagePath, err.Error())
		return
	}

	// If decoding succeeds, cache it.
	go cacheDomainFile(domain, imagePath, body)

	// If cache enabled, cache the file.
	imgHeaders(&w, r)

	if useFileCache {
		writeAndCacheImg(w, img, opts, getOptimizedImgCachePath(domain, imagePath, opts))
		return
	}

	// If cache disabled, just serve
	optimizer.Encode(w, img, opts)
	return

}

func respRemoteImgHandler(w http.ResponseWriter, r *http.Request) {

	var body []byte
	vars := mux.Vars(r)
	imagePath := vars["path"]
	domain := vars["domain"]

	// Set up options
	opts := optimizer.Options{
		Mime: mime.TypeByExtension(path.Ext(imagePath)),
	}

	opts.SetFromRequest(r)
	opts.Optimize()

	// Try to fetch cached iamge.
	if err := serveCachedRespImg(w, r, domain, imagePath, opts); err == nil {
		return
	}

	body, statusCode, err := getImageFromUpstream(domain, imagePath)
	if err != nil {
		imgHeaders(&w, r)
		http.Error(w, http.StatusText(statusCode), statusCode)
		return
	}

	optimizeAndServeImg(w, r, opts, domain, imagePath, body)
	return

}
