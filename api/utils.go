package api

import (
	"net/http"

	"github.com/golang/gddo/httputil/header"
)

func contentTypeIsJSON(h http.Header) bool {
	v, _ := header.ParseValueAndParams(h, "Content-Type")
	return v == "application/json"
}

func contentTypeAddJSON(h http.Header) {
	h.Add("Content-Type", "application/json")
}
