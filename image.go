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
	"errors"
	"github.com/chai2010/webp"
	log "gopkg.in/inconshreveable/log15.v2"
)

type Image struct {
	Path string
	Domain string
	Upstream string
	Mime string
	Dpr float64
	Width int
	StatusCode int
	Image image.Image
}

func (i *Image) Load() (error){

	// @todo Check for cached version
	err := i.loadFromCache()
	if err != nil {
		return err
	}

	// @todo If no cached version, get the image
	err = i.Get()
	if err != nil {
		return err
	}

	// @todo Process the image
	err = i.Process()

	return err

}

func (i *Image) Get() (err error) {

	// @todo Get the image from upstream server
	url := fmt.Sprintf("http://%s/%s", strings.Trim(i.Upstream, "/"), strings.Trim(i.Path, "/"))
	log.Info(fmt.Sprintf("Fetching %s", url))

	resp, err := http.Get(url)
  if err != nil {
		log.Error("Could not get %s: %s", url, err.Error())
      return
  }
	defer resp.Body.Close()

	// Make sure the request succeeded
	i.StatusCode = resp.StatusCode
	if i.StatusCode != 200 {
			err = errors.New(resp.Status)
			return
	}

	ctype := resp.Header.Get("Content-Type")
	if ! isRewritable(ctype) {
		err = errors.New("Unsupported format or not found")
		return
	}

	// Decode the data.
	newImg, _, err := image.Decode(resp.Body)
	if err != nil {
		log.Error("Could not decode %s: %s", url, err.Error())
		return
	}

	i.Image = newImg

	// Cache the original.
	i.CacheOriginal(ctype, i.Image)

	return

}

func (i *Image) Resize() {
	i.Image = resize.Resize(uint(float64(i.Width) * float64(i.Dpr)), 0, i.Image, resize.NearestNeighbor)
}

func (i *Image) Save() (err error) {
	return
}

func (i *Image) Log() (err error) {
	return
}

func (i *Image) Process() (err error) {

	// @todo Resize the image
	i.Resize()

	// @todo Cache the processed result
	i.Cache()

	// @todo Save the result
	i.Save()

	return

}

// @todo Check for cached version of image.
func (i *Image) loadFromCache() (err error) {
	return
}

// @todo Cache the image.
func (i *Image) Cache() (err error) {
	return
}

// @todo Cache the original image.
func (i *Image) CacheOriginal(mime string, img image.Image) (err error) {

	// case mime == "image/jpeg":
	// 	jpeg.Encode(w, i.Image, nil)
	// case mime == "image/png":
	// 	png.Encode(w, i.Image, nil)
	// case mime == "image/webp":
	// 	webp.Encode(w, i.Image, nil)
	// case mime == "image/gif":
	// 	gif.Encode(w, i.Image, nil)
	// }

	return
}

func (i *Image) getCacheKey() string {
	return fmt.Sprintf("%s--%s--%s@%s", i.getOrigCacheKey(), i.Mime, i.Width, i.Dpr)
}

func (i *Image) getOrigCacheKey() string {
	return fmt.Sprintf("%s/%s", i.Domain, i.Path)
}

func (i *Image) Write(w io.Writer) {

	if(i.Image == nil) {
		log.Error(fmt.Sprintf("Image is nil: %s", i))
		return
	}

	// Get the type.
	switch {
	case i.Mime == "image/jpeg":
		jpeg.Encode(w, i.Image, nil)
	case i.Mime == "image/png":
		png.Encode(w, i.Image)
	case i.Mime == "image/webp":
		webp.Encode(w, i.Image, nil)
	case i.Mime == "image/gif":
		gif.Encode(w, i.Image, nil)
	}

}
