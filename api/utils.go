package api

import (
	"encoding/json"
	"fmt"
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

func getParamStringFromURLAndBody(r *http.Request, param string) string {
	p := r.URL.Query().Get(param)
	if contentTypeIsJSON(r.Header) {
		b := make(map[string]string)
		err := json.NewDecoder(r.Body).Decode(&b)
		fmt.Println(b)
		if err == nil && b[param] != "" {
			p = b[param]
		}
	}
	return p
}
