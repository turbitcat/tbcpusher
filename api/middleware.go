package api

import (
	"net/http"

	"github.com/turbitcat/tbcpusher/v2/wsgo"
)

func requireString(keys ...string) wsgo.Handler {
	var f wsgo.Handler = func(c *wsgo.Context) {
		p := c.StringParams()
		for _, k := range keys {
			_, ok := p[k]
			if !ok {
				c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
				c.LogIfLogging("requireString: Requeir %v", k)
				return
			}
		}
		c.Next()
	}
	return f
}
