package main

import (
    "net/http"
    "github.com/gorilla/mux"
)

func RespImgHandler(w http.ResponseWriter, r *http.Request) {

    w.Header().Set("Accept-CH", "DPR, Width, Viewport-Width, Downlink")
    w.Header().Set("Vary", "Accept, DPR, Width, Save-Data, Downlink")

    vars := mux.Vars(r)

    path := vars["path"]
    domain := vars["domain"]

    i := Image{
        Domain: domain,
        Path: path,
    }

    // Try to fetch the image. If not, fail.
    statusCode, err := i.Get(); if err != nil {
        http.Error(w, err.Error(), statusCode)
        return
    }

    i.SetParamsFromRequest(r)

    // Set Content-Type and Cache-Control headers.
    w.Header().Set("Content-Type", i.Mime)

    // Serve the image.
    i.Write(w, r)

}
