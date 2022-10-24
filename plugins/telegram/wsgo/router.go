package wsgo

import (
	"net/http"
)

type handleLeaf struct {
	pattern string
	group
}

type group struct {
	parent   *group
	handlers []Handler
	root     *ServerMux
}

type ServerMux struct {
	getList    []handleLeaf
	postList   []handleLeaf
	handleList []handleLeaf
	handlers   []Handler
	notFound   *group
}

func (p *ServerMux) hit(s1 string, s2 string) bool {
	return s1 == s2
}

func appendReversly[T any](l []T, l2 []T) []T {
	for i := len(l2) - 1; i >= 0; i-- {
		l = append(l, l2[i])
	}
	return l
}

func (g *group) handle(w http.ResponseWriter, r *http.Request) {
	var handlers []Handler
	for p := g; p != nil; p = p.parent {
		handlers = appendReversly(handlers, p.handlers)
	}
	handlers = appendReversly(handlers, g.root.handlers)
	for i, j := 0, len(handlers)-1; i < j; i, j = i+1, j-1 {
		handlers[i], handlers[j] = handlers[j], handlers[i]
	}
	c := newContext()
	c.w = w
	c.r = r
	c.handlers = handlers
	handlers[0](c)
}

func (p *ServerMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var ml []handleLeaf
	switch r.Method {
	case http.MethodGet:
		ml = p.getList
	case http.MethodPost:
		ml = p.postList
	}
	for _, h := range ml {
		if p.hit(h.pattern, r.URL.Path) {
			h.group.handle(w, r)
			return
		}
	}
	for _, h := range p.handleList {
		if p.hit(h.pattern, r.URL.Path) {
			h.group.handle(w, r)
			return
		}
	}
	p.notFound.handle(w, r)
}

func (p *group) Group() *group {
	g := group{parent: p, root: p.root}
	return &g
}

func (p *group) Use(handler ...Handler) {
	p.handlers = append(p.handlers, handler...)
}

func (p *group) GET(pattern string, handler ...Handler) {
	leaf := handleLeaf{pattern, group{handlers: handler, parent: p, root: p.root}}
	p.root.getList = append(p.root.getList, leaf)
}

func (p *group) POST(pattern string, handler ...Handler) {
	leaf := handleLeaf{pattern, group{handlers: handler, parent: p, root: p.root}}
	p.root.postList = append(p.root.postList, leaf)
}

func (p *group) Handle(pattern string, handler ...Handler) {
	leaf := handleLeaf{pattern, group{handlers: handler, parent: p, root: p.root}}
	p.root.handleList = append(p.root.handleList, leaf)
}

func (p *ServerMux) Group() *group {
	g := group{parent: nil, root: p}
	return &g
}

func (p *ServerMux) Use(handler ...Handler) {
	p.handlers = append(p.handlers, handler...)
}

func (p *ServerMux) GET(pattern string, handler ...Handler) {
	leaf := handleLeaf{pattern, group{handlers: handler, parent: nil, root: p}}
	p.getList = append(p.getList, leaf)
}

func (p *ServerMux) POST(pattern string, handler ...Handler) {
	leaf := handleLeaf{pattern, group{handlers: handler, parent: nil, root: p}}
	p.postList = append(p.postList, leaf)
}

func (p *ServerMux) Handle(pattern string, handler ...Handler) {
	leaf := handleLeaf{pattern, group{handlers: handler, parent: nil, root: p}}
	p.handleList = append(p.handleList, leaf)
}

func (p *ServerMux) Run(addr string) error {
	return http.ListenAndServe(addr, p)
}
