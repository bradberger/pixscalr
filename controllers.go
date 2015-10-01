package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/bradberger/optimizer"
	"github.com/gorilla/mux"
	"github.com/mailgun/log"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/font"
	_ "golang.org/x/image/riff"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/vp8"
	_ "golang.org/x/image/vp8l"
	_ "golang.org/x/image/webp"
	_ "golang.org/x/image/webp/nycbcra"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

func isRewritable(ftype string) (rewrite bool) {

	accepted := map[string]bool{
		"image/webp":   true,
		"image/jpeg":   true,
		"image/pjpeg":  true,
		"image/png":    true,
		"image/gif":    true,
		"image/tiff":   true,
		"image/x-tiff": true,
	}

	if strings.HasPrefix(ftype, "image/") {
		ok, exist := accepted[ftype]
		if exist && ok {
			rewrite = true
		}
	}

	return

}

func addCacheHeader(w *http.ResponseWriter, str string) {
	(*w).Header().Set("X-Cached", str)
}

func addDurationHeader(w *http.ResponseWriter, start time.Time) {
	(*w).Header().Set("X-Duration", fmt.Sprintf("%v", time.Since(start)))
}

func cdnHeaders(w *http.ResponseWriter, r *http.Request) {
	disableCORSHeaders(w, r)
	(*w).Header().Set("X-Powered-By", fmt.Sprintf("GoFaster %s", version))
	(*w).Header().Set("Cache-Control:public", "max-age=31536000")
}

func disableCORSHeaders(w *http.ResponseWriter, r *http.Request) {
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

func respImgHandler(w http.ResponseWriter, r *http.Request) {

	var body []byte
	vars := mux.Vars(r)
	start := time.Now()
	imagePath := vars["path"]
	domain := vars["domain"]

	imgHeaders(&w, r)

	// Set up options
	opts := optimizer.Options{
		Mime: mime.TypeByExtension(path.Ext(imagePath)),
	}

	// Get the mime type.
	if strings.Contains(r.Header.Get("Accept"), "image/webp") {
		opts.Mime = "image/webp"
	}

	// Get the DPR
	dpr, err := strconv.ParseFloat(r.Header.Get("DPR"), 64)
	if err != nil {
		dpr, err = strconv.ParseFloat(r.FormValue("dpr"), 64)
	}
	if err != nil {
		dpr = 1.0
	}
	opts.Dpr = dpr

	// Set SaveData flag
	if r.Header.Get("Save-Data") == "1" || r.FormValue("save-data") == "1" {
		opts.SaveData = true
	} else {
		opts.SaveData = false
	}

	// Get the Viewport Width
	viewport, err := strconv.ParseFloat(r.Header.Get("Viewport-Width"), 64)
	if err != nil {
		viewport, err = strconv.ParseFloat(r.FormValue("viewport-width"), 64)
	}
	opts.ViewportWidth = viewport

	// Set the image width.
	width, err := strconv.Atoi(r.Header.Get("Width"))
	if err != nil {
		width, _ = strconv.Atoi(r.FormValue("width"))
	}
	if width > 0 {
		opts.Width = uint(width)
	}

	// Set Downlink
	downlink, err := strconv.ParseFloat(r.Header.Get("Downlink"), 64)
	if err != nil {
		downlink, err = strconv.ParseFloat(r.FormValue("downlink"), 64)
	}
	if err != nil {
		downlink = 0
	}
	opts.Downlink = downlink

	// Optimze the options to get permanent quality, etc.
	opts.Optimize()

	// Set Content-Type and Cache-Control headers.
	w.Header().Set("Content-Type", opts.Mime)

	// Try to fetch cached iamge.
	cacheExt := path.Ext(imagePath)
	if opts.Mime == "image/webp" {
		cacheExt = "webp"
	}

	cachePath := fmt.Sprintf("%s/%s/%s/%s--%vpx@%v--%v.%s", tmpDir, domain, path.Dir(imagePath), path.Base(imagePath), opts.Width, opts.Dpr, opts.Quality, cacheExt)
	if useFileCache {
		contents, err := ioutil.ReadFile(cachePath)
		if err == nil {
			log.Infof("HIT %s", cachePath)
			addCacheHeader(&w, "optimized")
			addDurationHeader(&w, start)
			w.Write(contents)
			return
		}

		// Check for resized original on disk.
		body, err = ioutil.ReadFile(fmt.Sprintf("%s/%s/%s", tmpDir, domain, imagePath))
		if err == nil {
			addCacheHeader(&w, "original")
		}

	}

	if len(body) == 0 {

		addCacheHeader(&w, "no")

		// Try to fetch the image. If not, fail.
		url := fmt.Sprintf("http://%s/%s", vars["domain"], imagePath)
		log.Infof("FETCH %s", url)

		resp, err := http.Get(url)
		if err != nil {
			log.Errorf("Couldn't get %s", url)
			http.Error(w, resp.Status, resp.StatusCode)
			return
		}
		defer resp.Body.Close()

		// Make sure the request succeeded
		if resp.StatusCode > 302 {
			log.Errorf("Couldn't fetch %s", url)
			http.Error(w, resp.Status, resp.StatusCode)
			return
		}

		// Get the mime from response.
		opts.Mime = resp.Header.Get("Content-Type")

		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("Couldn't read body %s", url)
			http.Error(w, http.StatusText(500), 500)
			return
		}

	}

	if useFileCache {
		// Cache the original response.
		go cacheFile(fmt.Sprintf("%s/%s/%s", tmpDir, vars["domain"], imagePath), body)
	}

	// Decode the image
	img, _, err := image.Decode(bytes.NewReader(body))
	if err != nil || img == nil {
		log.Errorf("Could not decode %s/%s: %s", vars["domain"], imagePath, err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}

	if isRewritable(opts.Mime){

		// Cache the file.
		if useFileCache {

			f, cacheWriter := imageCacheWriter(cachePath, img, opts)
			defer f.Close()

			multi := io.MultiWriter(w, cacheWriter)

			// Write image response.
			optimizer.Encode(multi, img, opts)
			cacheWriter.Flush()
			return

		}

		// If cache disabled, just serve
		addDurationHeader(&w, start)
		optimizer.Encode(w, img, opts)

		return

	}

	// Just serve it.
	addDurationHeader(&w, start)
	optimizer.Encode(w, img, opts)

}

func cacheFile(fileName string, data []byte) {

	dir := path.Dir(fileName)
	err := os.MkdirAll(dir, os.FileMode(0775))
	if err != nil {
		log.Errorf("Error caching image %s: %s", fileName, err)
		return
	}

	log.Infof("CACHE.PUT %s: %v bytes", fileName, len(data))
	ioutil.WriteFile(fileName, data, 0644)
	if err != nil {
		log.Errorf("Error caching image %s: %s", fileName, err)
		return
	}

}

// cacheOptimizedImageWriter Returns a file and a writer for the image resource.
// Be careful to flush the writer and close the file manually, as this function
// doesn't do that.
func imageCacheWriter(fileName string, img image.Image, opts optimizer.Options) (f *os.File, w *bufio.Writer) {

	dir := path.Dir(fileName)
	err := os.MkdirAll(dir, os.FileMode(0775))
	if err != nil {
		log.Errorf("Error caching image %s: %s", fileName, err)
		return
	}

	f, err = os.Create(fileName)
	if err != nil {
		log.Errorf("Error caching image %s: %s", fileName, err)
		return
	}

	w = bufio.NewWriter(f)

	return

}

func cdnHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	ext := path.Ext(r.URL.Path)

	f := File{
		Package:   vars["package"],
		Version:   vars["version"],
		Path:      vars["path"],
		Extension: ext,
		Mime:      mime.TypeByExtension(ext),
	}

	incoming := f.Query()
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(5 * time.Second)
		timeout <- true
	}()

	cdnHeaders(&w, r)
	w.Header().Set("Content-Type", f.Mime)

	select {
	case res := <-incoming:
		if res.Cached {
			log.Infof("CACHE %s", res.Path)
		} else {
			log.Infof("FETCH %s %s", res.Duration, res.Url)
		}
		w.Write(res.Bytes)
		return
	case <-timeout:
		http.Error(w, http.StatusText(504), 504)
		return
	}
}
