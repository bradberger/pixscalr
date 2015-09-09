package main

import (
    "net/http"
    "github.com/gorilla/mux"
)

func main() {

    r := mux.NewRouter()
    r.HandleFunc(`/{domain}/{path:[^?]+}`, RespImgHandler)

    http.ListenAndServe(":3000", r)

}
