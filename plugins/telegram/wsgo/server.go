package wsgo

func New() *ServerMux {
	s := ServerMux{}
	s.notFound = &group{root: &s, handlers: []Handler{NotFound}}
	return &s
}

func Default() *ServerMux {
	s := New()
	s.Use(Logger)
	s.Use(ParseParamsQuery)
	return s
}
