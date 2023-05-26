package wsgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/golang/gddo/httputil/header"
)

type Context struct {
	w            http.ResponseWriter
	r            *http.Request
	logs         []string
	logging      bool
	handlers     []Handler
	hdlI         int
	body         []byte
	rBodyerr     error
	params       map[string][]any
	stringParams map[string][]string
}

type firstTime struct{}

func (f *firstTime) Error() string {
	return "First time run"
}

func newContext() *Context {
	c := Context{}
	c.params = make(map[string][]any)
	c.stringParams = make(map[string][]string)
	c.rBodyerr = &firstTime{}
	return &c
}

func (c *Context) EnableLog() {
	c.logging = true
}

func (c *Context) DisableLog() {
	c.logging = false
}

func (c *Context) Log(format string, a ...any) {
	c.logs = append(c.logs, fmt.Sprintf(format, a...))
}

func (c *Context) log(format string, a ...any) {
	if c.logging {
		c.Log(format, a...)
	}
}

func (c *Context) Next() {
	c.hdlI++
	if c.hdlI < len(c.handlers) {
		c.handlers[c.hdlI](c)
	}
}

func (c *Context) GetResponseWriter() http.ResponseWriter {
	return c.w
}

func (c *Context) GetRequest() *http.Request {
	return c.r
}

func (c *Context) Query(p string) (string, bool) {
	return c.r.URL.Query().Get(p), c.r.URL.Query().Has(p)
}

func (c *Context) DefaultQuery(p string, defaultValue string) string {
	q := c.r.URL.Query()
	if q.Has(p) {
		return q.Get(p)
	}
	return defaultValue
}

func (c *Context) ReadAllBody() ([]byte, error) {
	if c.rBodyerr == nil {
		return c.body, nil
	}
	var ft *firstTime
	if errors.As(c.rBodyerr, &ft) {
		if c.body, c.rBodyerr = io.ReadAll(c.r.Body); c.rBodyerr != nil {
			return nil, c.rBodyerr
		} else {
			return c.body, c.rBodyerr
		}
	} else {
		return nil, c.rBodyerr
	}
}

func (c *Context) ContentType() string {
	v, _ := header.ParseValueAndParams(c.r.Header, "Content-Type")
	return v
}

func (c *Context) BindJSON(v any) error {
	b, error := c.ReadAllBody()
	if error != nil {
		return error
	}
	return json.Unmarshal(b, v)
}

func (c *Context) AddParam(key string, v any) {
	c.params[key] = append(c.params[key], v)
	s, ok := v.(string)
	if ok {
		c.stringParams[key] = append(c.stringParams[key], s)
	}
}

func (c *Context) Param(k string) (any, bool) {
	v, ok := c.params[k]
	if !ok {
		return nil, false
	}
	if len(v) == 0 {
		return nil, true
	}
	return v[len(v)-1], true
}

func (c *Context) DefaultParam(k string, d any) any {
	v, ok := c.Param(k)
	if !ok {
		return d
	}
	return v
}

func (c *Context) StringParam(k string) (string, bool) {
	v, ok := c.stringParams[k]
	if !ok {
		return "", false
	}
	if len(v) == 0 {
		return "", true
	}
	return v[len(v)-1], true
}

func (c *Context) DefaultStringParam(k string, d string) string {
	v, ok := c.StringParam(k)
	if !ok {
		return d
	}
	return v
}

func (c *Context) Params() map[string]any {
	d := map[string]any{}
	for k, vs := range c.params {
		if len(vs) == 0 {
			d[k] = nil
		} else {
			d[k] = vs[len(vs)-1]
		}
	}
	return d
}

func (c *Context) StringParams() map[string]string {
	d := map[string]string{}
	for k, vs := range c.stringParams {
		if len(vs) == 0 {
			d[k] = ""
		} else {
			d[k] = vs[len(vs)-1]
		}
	}
	return d
}

func (c *Context) AllParams() map[string][]any {
	return c.params
}

func (c *Context) AllStringParams() map[string][]string {
	return c.stringParams
}

func (c *Context) StatusCode(code int) {
	c.w.WriteHeader(code)
	c.log("StatusCode [%d]", code)
}

func (c *Context) String(code int, s string) {
	c.w.Header().Set("Content-Type", "text/plain;charset=UTF-8")
	c.w.WriteHeader(code)
	io.WriteString(c.w, s)
	c.log("String [%d]: %v", code, s)
}

func (c *Context) Json(code int, data any) {
	c.w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	c.w.WriteHeader(code)

	json.NewEncoder(c.w).Encode(data)
	c.log("Json [%d]: %v", code, data)
}

func (c *Context) Stream(code int, data []byte) {
	c.w.Header().Set("Content-Type", "application/octet-stream")
	c.w.WriteHeader(code)
	c.w.Write(data)
	c.log("Stream [%d]", code)
}
