package api

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/golang/gddo/httputil/header"
)

func contentTypeIsJSON(h http.Header) bool {
	v, _ := header.ParseValueAndParams(h, "Content-Type")
	return v == "application/json"
}

func contentTypeAddJSON(h http.Header) {
	h.Add("Content-Type", "application/json")
}

func parseStringParamsFromJSON(data []byte) (url.Values, error) {
	var t map[string]string
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	ret := url.Values{}
	for k, v := range t {
		ret[k] = append(ret[k], v)
	}
	return ret, nil
}

func valuesUnion(a url.Values, b url.Values) url.Values {
	keys := map[string]struct{}{}
	for k := range a {
		keys[k] = struct{}{}
	}
	for k := range b {
		keys[k] = struct{}{}
	}
	r := url.Values{}
	for k := range keys {
		r[k] = append(a[k], b[k]...)
	}
	return r
}
