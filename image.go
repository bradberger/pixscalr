package main

import (
	"fmt"
	"io"
	"net/http"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"image/png"
	"image/gif"
	"strings"
	"strconv"
	"errors"
	"github.com/chai2010/webp"
	log "gopkg.in/inconshreveable/log15.v2"
)

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

type Image struct {
    Mime string
    Domain string
    Path string
    Width uint
	Dpr float64
	Quality int
	Downlink float64
	SaveData bool
	ViewportWidth int
	Image image.Image
}

func (i *Image) Get() (statusCode int, err error) {

	url := fmt.Sprintf("http://%s/%s", i.Domain, i.Path)
	log.Info(fmt.Sprintf("Fetching %s", url))

	resp, err := http.Get(url)
	if err != nil {
		statusCode = 500
		log.Error("Could not get %s: %s", url, err.Error())
		return
	}
	defer resp.Body.Close()

	// Make sure the request succeeded
	statusCode = resp.StatusCode
	if statusCode > 302 {
		log.Error(fmt.Sprintf("Couldn't fetch %s", url))
		err = errors.New(resp.Status)
		return
	}

	i.Mime = resp.Header.Get("Content-Type")
	if ! isRewritable(i.Mime) {
		err = errors.New("Unsupported format or not found")
		return
	}

	// Decode the data.
	i.Image, _, err = image.Decode(resp.Body)
	if err != nil || i.Image == nil {
		log.Error(fmt.Sprintf("Could not decode %s: %s", url, err.Error()))
		return
	}

	// Set the width now.
	i.Width = uint(i.Image.Bounds().Max.X)

	log.Info(fmt.Sprintf("Decoded image from %s with original width of %v", url, i.Width))


	return

}

// Now check for WebP
func (i *Image) SetWebP(r *http.Request) {
	if strings.Contains(r.Header.Get("Accept"), "image/webp") {
		log.Info("WebP here we come")
		i.Mime = "image/webp"
	}
}

func (i *Image) SetParamsFromRequest(r *http.Request) {
	i.SetWebP(r)
    i.SetSaveDataFromRequest(r)
    i.SetDprFromRequest(r)
	i.SetViewportWidthFromRequest(r)
    i.SetWidthFromRequest(r)
    i.SetDownlinkFromRequest(r)
}

func (i *Image) SetWidthFromRequest(r *http.Request) {

	w, err := strconv.Atoi(r.Header.Get("Width"))
	if err != nil {
		w, _ = strconv.Atoi(r.FormValue("width"))
	}

	// If one of two previous actions didn't return error, set the width.
	if w > 0 {
		i.Width = uint(w)
	}

	// Make sure the image isn't bigger than viewport-width
	log.Info(fmt.Sprintf("Width: %v Viewport-Width: %v", i.Width, i.ViewportWidth))
	if i.ViewportWidth > 0 {
		if i.Width > uint(i.ViewportWidth) {
			i.Width = uint(i.ViewportWidth)
		}
	}

}

func (i *Image) SetDownlinkFromRequest(r *http.Request) {

	// @see https://en.wikipedia.org/wiki/Comparison_of_wireless_data_standards
	// 0.384Mbps (GPRS EDGE)
	downlink, err := strconv.ParseFloat(r.Header.Get("Downlink"), 64)
	if err != nil {
		downlink, err = strconv.ParseFloat(r.FormValue("downlink"), 64)
	}

	if err != nil {
		downlink = 0
	}

	i.Downlink = downlink

}

// @see https://github.com/rflynn/imgmin#quality-details
func (i *Image) SetQuality() {

	// Set quality dependant on the DPR.
	i.Quality = int(100 - i.Dpr * 30)

	// @todo @research Set quality based on Downlink
	switch {
	case i.Downlink > 0 && i.Downlink < 1:
		i.Quality = int(float32(i.Quality) * float32(i.Downlink))
		break
	}

	// @todo @research Set quality based on SaveData
	if i.SaveData {
		i.Quality = int(float32(i.Quality) * float32(0.75))
	}

}

func (i *Image) SetDprFromRequest(r *http.Request) {

	dpr, err := strconv.ParseFloat(r.Header.Get("DPR"), 64)
	if err != nil {
		dpr, err = strconv.ParseFloat(r.FormValue("dpr"), 64)
	}

	if err != nil {
		dpr = 1.0
	}

	i.Dpr = dpr

}

func (i *Image) SetViewportWidthFromRequest(r *http.Request) {

	w, err := strconv.Atoi(r.Header.Get("Viewport-Width"))
	if err != nil {
		w, err = strconv.Atoi(r.FormValue("viewport-width"))
	}

	log.Info(fmt.Sprintf("Setting viewport width as %v", w))

	i.ViewportWidth = w

}

func (i *Image) SetSaveDataFromRequest(r *http.Request) {
	if r.Header.Get("Save-Data") == "1" || r.FormValue("save-data") == "1" {
		i.SaveData = true
	} else {
		i.SaveData = false
	}
}

func (i *Image) Resize() {
	if i.Width != 0 {
		i.Image = resize.Resize(i.Width, 0, i.Image, resize.NearestNeighbor)
	}
}

func (i *Image) Write(w io.Writer, r *http.Request) {

	// Handle WebP, DPR, Quality, Width, and Size
	i.SetQuality()
	i.Resize()

	log.Info(fmt.Sprintf("Quality of %s/%s is %v", i.Domain, i.Path, i.Quality))

	// Now encode the image according to the type.
	switch {
	case i.Mime == "image/jpeg":
		jpeg.Encode(w, i.Image, &jpeg.Options{ Quality: i.Quality })
	case i.Mime == "image/png":
		// @todo Set the compression level
		png.Encode(w, i.Image)
	case i.Mime == "image/webp":
		webp.Encode(w, i.Image, &webp.Options{ Quality: float32(i.Quality) })
	case i.Mime == "image/gif":
		gif.Encode(w, i.Image, nil)
	}

}
