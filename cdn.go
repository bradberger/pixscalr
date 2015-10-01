package main

import (
    "fmt"
    "net/http"
    "io/ioutil"
    "strings"
    "github.com/mailgun/log"
    "time"
    "path"
    "os"
)

// File is a struct which contains necessary information regarding a static file.
// The related methods allow for fetching and serving that file via various CDN's.
type File struct {
    Package string
    Version string
    Path string
    Extension string
    Mime string
}

// FileQueryResult holds data related to a file lookup.
// This includes where it came from, it's size, contents, etc.
type FileQueryResult struct {
    StausCode int
    Cached bool
    Url string
    Path string
    Bytes []byte
    Size int
    Duration time.Duration
}

func (f *File) getCdnjsURL() (url string) {
    if f.Package == "" || f.Version == "" || f.Path == "" {
        url = ""
        return
    }
    url = fmt.Sprintf("https://cdnjs.cloudflare.com/ajax/libs/%s/%s/%s", f.Package, f.Version, f.Path)
    return
}

func (f *File) getJsDelivrURL() (url string) {
    if f.Package == "" || f.Version == "" || f.Path == "" {
        url = ""
        return
    }
    url = fmt.Sprintf("https://cdn.jsdelivr.net/%s/%s/%s", f.Package, f.Version, f.Path)
    return
}

func (f *File) getGoogleApisURL() (url string) {
    if f.Package == "" || f.Version == "" || f.Path == "" {
        url = ""
        return
    }
    url = fmt.Sprintf("https://ajax.googleapis.com/ajax/libs/%s/%s/%s", f.Package, f.Version, f.Path)
    return
}

func (f *File) getAspNetURL() (url string) {
    if f.Package == "" || f.Version == "" || f.Path == "" {
        url = ""
        return
    }
    path := strings.TrimRight(f.Path, f.Extension)
    min :=  strings.HasSuffix(path, ".min")
    minStr := ""
    if min {
        path = strings.TrimRight(path, ".min")
        minStr = ".min"
    }

    url = fmt.Sprintf("https://ajax.aspnetcdn.com/ajax/%s/%s-%s%s%s", f.Package, path, f.Version, minStr, f.Extension)
    return
}

func (f *File) getYandexURL() (url string) {
    if f.Package == "" || f.Version == "" || f.Path == "" {
        url = ""
        return
    }
    url = fmt.Sprintf("https://yastatic.net/%s/%s/%s", f.Package, f.Version, f.Path)
    return
}

func (f *File) getOssCdnURL() (url string) {
    if f.Package == "" || f.Version == "" || f.Path == "" {
        url = ""
        return
    }
    url = fmt.Sprintf("https://oss.maxcdn.com/%s/%s/%s", f.Package, f.Version, f.Path)
    return
}

// GetUrls returns a set of possible locations for each file based on available CDNs.
func (f *File) GetUrls() []string {
    return []string{
        f.getCdnjsURL(),
        f.getJsDelivrURL(),
        f.getGoogleApisURL(),
        f.getOssCdnURL(),
        f.getYandexURL(),
        f.getAspNetURL(),
    }
}

func (f *File) getCachePath() (path string) {
    path = fmt.Sprintf("%s/cdn/%s/%s/%s", tmpDir, f.Package, f.Version, f.Path)
    return
}

func getUrl(url string) <-chan FileQueryResult {

    out := make(chan FileQueryResult, 1)
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
        out <- FileQueryResult{
            Cached: false,
            Url: url,
            Bytes: body,
            Size: len(body),
            Duration: elapsed,
        }

        log.Infof("DONE %s %s", elapsed, url)

    }()

    return out

}

// CacheToDisk writes the content to disk.
func (f *File) CacheToDisk(content []byte) (err error) {

    if len(content) < 0 {
        return
    }

    fn := f.getCachePath()
    dir := path.Dir(fn)

    // Create the directory.
    err = os.MkdirAll(dir, os.FileMode(0775))
    if err != nil {
        return
    }

    // Write the file
    err = ioutil.WriteFile(fn, content, 0644)
    return

}

// Query returns the file contents in []byte for each file.
// It loops through the possible file URL's, and returns the first result.
// URL's which have errors, don't exist, etc., block forever.
func (f *File) Query() <-chan FileQueryResult {

    urls := f.GetUrls()
    out := make(chan FileQueryResult, len(urls))

    go func() {
        contents, err := ioutil.ReadFile(f.getCachePath())
        if err == nil {
            out <- FileQueryResult{
                Cached: true,
                Path: f.getCachePath(),
                Bytes: contents,
                Size: len(contents),
            }
            close(out)
        } else {
            haveResult := false
            for _, url := range urls {
                go func(uri string) {
                    r := <-getUrl(uri)

                    if ! haveResult {
                        // First hit, cache it.
                        haveResult = true
                        if err := f.CacheToDisk(r.Bytes); err != nil {
                            log.Errorf("Failed to cache file %s: %s", f.getCachePath(), err)
                        }

                    }

                    out <- r
                }(url)
            }
        }
    }()

    return out

}
