package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mailgun/log"
	"io/ioutil"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"
)

// CDNFile is a struct which contains necessary information regarding a static file.
// The related methods allow for fetching and serving that file via various CDN's.
type CDNFile struct {
	Package   string
	Version   string
	Path      string
	Extension string
	Mime      string
}

// CDNQuery holds data related to a file lookup.
// This includes where it came from, it's size, contents, etc.
type CDNQuery struct {
	StausCode int
	Cached    bool
	URL       string
	Path      string
	Bytes     []byte
	Size      int
	Duration  time.Duration
}

func (f *CDNFile) getCdnjsURL() (url string) {
	if f.Package == "" || f.Version == "" || f.Path == "" {
		url = ""
		return
	}
	url = fmt.Sprintf("https://cdnjs.cloudflare.com/ajax/libs/%s/%s/%s", f.Package, f.Version, f.Path)
	return
}

func (f *CDNFile) getJsDelivrURL() (url string) {
	if f.Package == "" || f.Version == "" || f.Path == "" {
		url = ""
		return
	}
	url = fmt.Sprintf("https://cdn.jsdelivr.net/%s/%s/%s", f.Package, f.Version, f.Path)
	return
}

func (f *CDNFile) getGoogleApisURL() (url string) {
	if f.Package == "" || f.Version == "" || f.Path == "" {
		url = ""
		return
	}
	url = fmt.Sprintf("https://ajax.googleapis.com/ajax/libs/%s/%s/%s", f.Package, f.Version, f.Path)
	return
}

func (f *CDNFile) getAspNetURL() (url string) {
	if f.Package == "" || f.Version == "" || f.Path == "" {
		url = ""
		return
	}
	path := strings.TrimRight(f.Path, f.Extension)
	min := strings.HasSuffix(path, ".min")
	minStr := ""
	if min {
		path = strings.TrimRight(path, ".min")
		minStr = ".min"
	}

	url = fmt.Sprintf("https://ajax.aspnetcdn.com/ajax/%s/%s-%s%s%s", f.Package, path, f.Version, minStr, f.Extension)
	return
}

func (f *CDNFile) getYandexURL() (url string) {
	if f.Package == "" || f.Version == "" || f.Path == "" {
		url = ""
		return
	}
	url = fmt.Sprintf("https://yastatic.net/%s/%s/%s", f.Package, f.Version, f.Path)
	return
}

func (f *CDNFile) getOssCdnURL() (url string) {
	if f.Package == "" || f.Version == "" || f.Path == "" {
		url = ""
		return
	}
	url = fmt.Sprintf("https://oss.maxcdn.com/%s/%s/%s", f.Package, f.Version, f.Path)
	return
}

// GetUrls returns a set of possible locations for each file based on available CDNs.
func (f *CDNFile) GetUrls() []string {
	return []string{
		f.getCdnjsURL(),
		f.getJsDelivrURL(),
		f.getGoogleApisURL(),
		f.getOssCdnURL(),
		f.getYandexURL(),
		f.getAspNetURL(),
	}
}

func (f *CDNFile) getCachePath() (path string) {
	path = fmt.Sprintf("%s/cdn/%s/%s/%s", tmpDir, f.Package, f.Version, f.Path)
	return
}

func getURL(url string) <-chan CDNQuery {

	out := make(chan CDNQuery, 1)
	go func() {

		if url == "" {
			return
		}

		log.Infof("GET %s", url)

		start := time.Now()

		// @todo Keep a list of 404's/errors and don't look them up.
		resp, err := http.Get(url)
		if err != nil || resp.StatusCode != 200 {
			log.Errorf("ERROR %v %s", resp.StatusCode, url)
			return
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("ERROR %s %s", err.Error(), url)
			return
		}

		elapsed := time.Since(start)
		out <- CDNQuery{
			Cached:   false,
			URL:      url,
			Bytes:    body,
			Size:     len(body),
			Duration: elapsed,
		}

		log.Infof("DONE %s %s", elapsed, url)

	}()

	return out

}

// CacheToDisk writes the content to disk.
func (f *CDNFile) CacheToDisk(content []byte) {
	cacheFile(f.getCachePath(), content)
}

// Query returns the file contents in []byte for each file.
// It loops through the possible file URL's, and returns the first result.
// URL's which have errors, don't exist, etc., block forever.
func (f *CDNFile) Query() <-chan CDNQuery {

	urls := f.GetUrls()
	out := make(chan CDNQuery, len(urls))

	go func() {
		contents, err := ioutil.ReadFile(path.Clean(f.getCachePath()))
		if err == nil {
			out <- CDNQuery{
				Cached: true,
				Path:   f.getCachePath(),
				Bytes:  contents,
				Size:   len(contents),
			}
			close(out)
		} else {
			haveResult := false
			for _, url := range urls {
				go func(uri string) {
					r := <-getURL(uri)
					if !haveResult {
						// First hit, cache it.
						haveResult = true
						f.CacheToDisk(r.Bytes)
					}
					out <- r
				}(url)
			}
		}
	}()

	return out

}

func cdnHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	ext := path.Ext(r.URL.Path)

	f := CDNFile{
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
			log.Infof("FETCH %s %s", res.Duration, res.URL)
		}
		w.Write(res.Bytes)
		return
	case <-timeout:
		http.Error(w, http.StatusText(504), 504)
		return
	}
}

func cdnHeaders(w *http.ResponseWriter, r *http.Request) {
	disableCORSHeaders(w, r)
	(*w).Header().Set("X-Powered-By", fmt.Sprintf("GoFaster %s", version))
	(*w).Header().Set("Cache-Control:public", "max-age=31536000")
}
