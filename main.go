package main

import (
	_ "expvar"
	"flag"
	"fmt"
	"github.com/bradberger/gocdn"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/mailgun/log"
	"net/http"
	"os"
)

var documentRoot string
var useFileCache bool
var tmpDir string
var version string
var serverName string

func init() {

	version = "0.1.0"
	serverName = "pixscalr"

	// Set up the logger.
	console, _ := log.NewLogger(log.Config{"console", "info"})
	syslog, _ := log.NewLogger(log.Config{"syslog", "error"})
	log.Init(console, syslog)

}

func main() {

	// Set up command line flags.
	addr := flag.String("listen", ":3000", "IP address/port to listen on")
	cacheDir := flag.String("cache-dir", fmt.Sprintf("%s/%s", os.TempDir(), serverName), "Directory for caching static resources")
	disableCdn := flag.Bool("disable-cdn", false, "Disable the CDN functionality")
	disableProxy := flag.Bool("disable-proxy", false, "Disable the CDN functionality")
	cdnPrefix := flag.String("prefix-cdn", "/cdn/", "The prefix for the fetching CDN resources")
	proxyPrefix := flag.String("prefix-proxy", "/assets/", "The prefix for handling GET proxy requests")
	debug := flag.Bool("debug", true, "Publish stats to /debug/vars")

	flag.StringVar(&documentRoot, "docroot", "", "Document root to serve static files. Default is none (disabled)")
	flag.BoolVar(&useFileCache, "cache", true, "Enable filesystem caching")
	flag.Parse()

	tmpDir = *cacheDir

	// Make sure cache directory exists.
	// and check if we can write to file cache.
	initCacheDir()

	r := mux.NewRouter()

	// CDN proxy subrouter
	if !*disableCdn && len(*cdnPrefix) > 0 {

		cdnCacheDir := fmt.Sprintf("%s/libraries", tmpDir)
		cdnRouter := r.PathPrefix(*cdnPrefix).Subrouter()

		log.Infof("Serving cdn files from %s%s", *addr, *cdnPrefix)
		if useFileCache {
			log.Infof("Caching cdn files in %s", cdnCacheDir)
		}

		cdn := gocdn.CDN{
			Prefix:        *cdnPrefix,
			CacheDuration: 100,
			Cors:          true,
			CacheDir:      cdnCacheDir,
			UseFileCache:  useFileCache,
		}

		cdnRouter.HandleFunc(`/{package}/{version}/{path:[^?]+}`, cdn.Handler)

	}

	if !*disableProxy && len(*proxyPrefix) > 0 {
		// General proxy subrouter.
		log.Infof("Serving assets via proxy from %s%s ", *addr, *proxyPrefix)
		proxyRouter := r.PathPrefix(*proxyPrefix).Subrouter()
		proxyRouter.HandleFunc(`/{domain}/{path:.+(jpe?g|png|gif|webp|tiff|bmp)}`, respRemoteImgHandler)
		proxyRouter.HandleFunc(`/{domain}/{path:[^?]+}`, proxyHandler)
	}

	if documentRoot != "" {
		log.Infof("Serving static files from %s", documentRoot)
		r.HandleFunc(`/{.+(jpe?g|png|gif|webp|tiff|bmp)}`, respLocalImgHandler)
		r.PathPrefix("/").Handler(http.FileServer(http.Dir(documentRoot)))
		http.Handle("/", r)
	}

	if *debug {
		log.Infof("Publishing stats to /debug/vars")
		r.Handle("/debug/vars", http.DefaultServeMux)
	}

	log.Infof("Ready to listen on %s", *addr)
	log.Errorf("Server died: %s", http.ListenAndServe(*addr, handlers.CompressHandler(r)))

}

func disableCORSHeaders(w *http.ResponseWriter, r *http.Request) {
	if origin := r.Header.Get("Origin"); origin != "" {
		(*w).Header().Set("Access-Control-Allow-Origin", origin)
		(*w).Header().Set("Timing-Allow-Origin", origin)
		(*w).Header().Set("Access-Control-Allow-Methods", r.Method)
		(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
	}
}
