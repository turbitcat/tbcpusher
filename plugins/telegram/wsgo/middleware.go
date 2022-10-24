package wsgo

import (
	"fmt"
	"net/http"
	"time"
)

func Logger(c *Context) {
	t := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%v] %v %v\n", t, c.r.Method, c.r.URL.Path)
	c.Next()
	for _, log := range c.logs {
		fmt.Printf("                      - %v\n", firstN(log, 512))
	}
}

func ParseParamsQuery(c *Context) {
	for k, ql := range c.r.URL.Query() {
		for _, q := range ql {
			c.AddParam(k, q)
		}
	}
	c.Next()
}

func ParseParamsJSON(c *Context) {
	defer c.Next()
	if c.ContentType() != "application/json" {
		return
	}
	dir := map[string]any{}
	if err := c.BindJSON(&dir); err != nil {
		return
	}
	for k, v := range dir {
		c.AddParam(k, v)
	}
}

func NotFound(c *Context) {
	http.NotFound(c.w, c.r)
	c.log("NotFound [404]")
}
