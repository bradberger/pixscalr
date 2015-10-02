package main

import (
	"bufio"
	"fmt"
	"github.com/bradberger/optimizer"
	"github.com/mailgun/log"
	"image"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

func initCacheDir() {
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
		if useFileCache {
			log.Infof("Caching via filesystem enabled")
			log.Infof("Using %v as cache path", tmpDir)
		}
	}
}

// cacheOptimizedImageWriter Returns a file and a writer for the image resource.
// Be careful to flush the writer and close the file manually, as this function
// doesn't do that.
func fileCacheWriter(fileName string) (f *os.File, w *bufio.Writer) {

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

func writeAndCacheImg(w io.Writer, img image.Image, opts optimizer.Options, fileName string) {

	f, cacheWriter := fileCacheWriter(fileName)
	defer f.Close()

	multi := io.MultiWriter(w, cacheWriter)

	// Write image response.
	optimizer.Encode(multi, img, opts)
	cacheWriter.Flush()
	return

}

func writeAndCache(w io.Writer, r io.Reader, fileName string) {

	f, cacheWriter := fileCacheWriter(fileName)
	defer f.Close()

	multi := io.MultiWriter(w, cacheWriter)
	io.Copy(multi, r)

}

func cacheFile(fileName string, data []byte) {
	if useFileCache && len(data) > 0 {

		go func(fileName string, data []byte) {

			fileName = path.Clean(fileName)
			dir := path.Dir(fileName)

			if err := os.MkdirAll(dir, os.FileMode(0775)); err != nil {
				log.Errorf("Error caching image %s: %s", fileName, err)
				return
			}

			if err := ioutil.WriteFile(fileName, data, 0644); err != nil {
				log.Errorf("Error caching image %s: %s", fileName, err)
				return
			}

		}(fileName, data)

	}
}

func cacheTmpFile(fileName string, data []byte) {
	cacheFile(fmt.Sprintf("%s/%s", tmpDir, fileName), data)
}

func cacheDomainFile(domain string, filePath string, data []byte) {
	cacheTmpFile(fmt.Sprintf("%s/%s", domain, filePath), data)
}

func getOptimizedImgCachePath(domain string, imagePath string, opts optimizer.Options) string {

	cacheExt := path.Ext(imagePath)
	if opts.Mime == "image/webp" {
		cacheExt = ".webp"
	}

	return fmt.Sprintf("%s/%s/%s/%s--%vpx@%v--%v%s", tmpDir, domain, path.Dir(imagePath), path.Base(imagePath), opts.Width, opts.Dpr, opts.Quality, cacheExt)
}

func getImageFromCache(imagePath string) (body []byte, err error) {
	body, err = ioutil.ReadFile(path.Clean(imagePath))
	if err != nil && len(body) == 0 {
		err = fmt.Errorf("Empty image file %s")
	}
	return
}

func addCacheHeader(w *http.ResponseWriter, str string) {
	(*w).Header().Set("X-Cached", str)
}
