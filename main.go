package main

import (
  "fmt"
  "mime"
  "strings"
  "net/http"
  "strconv"
  log "gopkg.in/inconshreveable/log15.v2"
)

func main() {

  http.HandleFunc("/", serveImgCtrl)

  log.Info("Starting server on :3000")
  log.Crit(fmt.Sprintf("Server died %s", http.ListenAndServe(":3000", nil)))

}

func serveImgCtrl(w http.ResponseWriter, r *http.Request) {

  log.Info(fmt.Sprintf("%s", r.URL.Path))

  ftype := getMimeFromReq(r)
  if ! isRewritable(ftype) {
    http.Error(w, "Mime type not supported", http.StatusUnsupportedMediaType)
		return
  }

  err, domain := getDomainFromReq(r)
  if err != nil {
    http.Error(w, "Invalid domain", http.StatusPreconditionFailed)
		return
  }

  if(acceptsWebP(r)) {
    ftype = "image/webp"
  }

  dpr := getDPRFromReq(r)

  img := Image{
    Mime: ftype,
    Dpr: dpr,
    Path: r.URL.Path,
    Domain: domain.Domain,
    Upstream: domain.Upstream,
    Width: getWidthFromReq(r),
  }

  // @todo Get the image
  // @todo Handle failing to get the image
  err = img.Load(); if err != nil {

    statusCode := 500
    if (img.StatusCode != 0) {
      statusCode = img.StatusCode
    }

    http.Error(w, err.Error(), statusCode)
    return
  }

  defer img.Log()

  w.Header().Set("Content-Type", img.Mime)
  w.Header().Set("Content-DPR", fmt.Sprintf("%v", dpr))

  img.Write(w)

  return

}

func getWidthFromReq(r *http.Request) (w int) {

  w, err := strconv.Atoi(r.Header.Get("Width"))
  if err != nil {
    w, err = strconv.Atoi(r.FormValue("width"))
  }

  if err != nil {
    w = 0
  }

  return

}

func getDPRFromReq(r *http.Request) (float64) {

  dpr, err := strconv.ParseFloat(r.Header.Get("DPR"), 64)

  if err != nil {
    dpr, err = strconv.ParseFloat(r.FormValue("dpr"), 64)
  }

  if err != nil {
    dpr = 1.0
  }

  return dpr

}

func isRewritable(ftype string) (rewrite bool) {

  accepted := map[string]bool{
      "image/webp": true,
      "image/jpeg": true,
      "image/pjpeg": true,
      "image/png": true,
      "image/gif": true,
      "image/tiff": true,
      "image/x-tiff": true,
  }

  if (strings.HasPrefix(ftype, "image/")) {

    ok, exist := accepted[ftype]
    if(exist && ok) {
      rewrite = true
    }

  }

  return

}

func acceptsWebP(r *http.Request) bool {
  return strings.Contains(r.Header.Get("Accept"), "image/webp")
}

func getMimeFromReq(r *http.Request) string {

  parts := strings.Split(r.URL.Path, ".")

  return mime.TypeByExtension("." + parts[len(parts)-1])

}