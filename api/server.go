package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/turbitcat/tbcpusher/v2/database"
)

// Article ...
type Article struct {
	Title  string `json:"Title"`
	Author string `json:"author"`
	Link   string `json:"link"`
}

// Articles ...
var Articles []Article = []Article{
	{Title: "Python Intermediate and Advanced 101",
		Author: "Arkaprabha Majumdar",
		Link:   "https://www.amazon.com/dp/B089KVK23P"},
	{Title: "R programming Advanced",
		Author: "Arkaprabha Majumdar",
		Link:   "https://www.amazon.com/dp/B089WH12CR"},
	{Title: "R programming Fundamentals",
		Author: "Arkaprabha Majumdar",
		Link:   "https://www.amazon.com/dp/B089S58WWG"},
}

type Server struct {
	db      database.Database
	addr    string
	prefix  string
	checkCT bool
}

const pathCreateGroup = "/group/create"
const pathPushToGroup = "/group/push"
const pathCreateSession = "/session/create"
const pathPushToSession = "/session/push"
const pathCheckSession = "/session/check"
const pathHideSession = "/session/hide"

func NewServer(db database.Database) Server {
	return Server{db: db, addr: ":8000"}
}

func (s *Server) SetAddr(addr string) {
	s.addr = addr
}

func (s *Server) SetPrefix(p string) {
	if p != "" && p[0] != '/' {
		p = "/" + p
	}
	s.prefix = p
}

func (s *Server) SetContenetTypeCheck(b bool) {
	s.checkCT = b
}

func (s *Server) Serve() error {
	s.HandleFunc(s.prefix+pathCreateGroup, s.createGroup)
	s.HandleFunc(s.prefix+pathPushToGroup, s.pushToGroup)
	s.HandleFunc(s.prefix+pathCreateSession, s.createSession)
	s.HandleFunc(s.prefix+pathPushToSession, s.pushToSession)
	s.HandleFunc(s.prefix+pathCheckSession, s.checkSession)
	s.HandleFunc(s.prefix+pathHideSession, s.hideSession)
	return http.ListenAndServe(s.addr, nil)
}

type ProcessedRequest struct {
	*http.Request
	BodyBytes []byte
	Params    url.Values
}

func (s *Server) processR(r *http.Request) (*ProcessedRequest, error) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("processR: %v", err)
	}
	var pr ProcessedRequest
	pr.Request = r
	pr.BodyBytes = b
	p := r.URL.Query()
	if !s.checkCT || contentTypeIsJSON(r.Header) {
		if bP, err := parseStringParamsFromJSON(b); err == nil {
			p = valuesUnion(bP, p)
		}
	}
	pr.Params = p
	return &pr, nil
}

func (s *Server) HandleFunc(pattern string, handler func(http.ResponseWriter, *ProcessedRequest)) {
	wrapper := func(w http.ResponseWriter, r *http.Request) {
		rr, err := s.processR(r)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			print(err.Error())
		}
		handler(w, rr)
	}
	http.HandleFunc(pattern, wrapper)
}

// info={}
func (s *Server) createGroup(w http.ResponseWriter, r *ProcessedRequest) {
	fmt.Println("Endpoint Hit: createGroup")
	info := r.Params.Get("info")
	id, err := s.db.NewGroup(info)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		print(err.Error())
		return
	}
	contentTypeAddJSON(w.Header())
	json.NewEncoder(w).Encode(map[string]any{"id": id})
}

// group={groupid}&hook={callbackurl}&info={}
func (s *Server) createSession(w http.ResponseWriter, r *ProcessedRequest) {
	fmt.Println("Endpoint Hit: createSession")
	groupID := r.Params.Get("group")
	g, err := s.db.GetGroupByID(groupID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		print(err.Error())
		return
	}
	hook := r.Params.Get("hook")
	if hook == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		print("hook url is empty")
		return
	}
	info := r.Params.Get("info")
	id, err := g.NewSession(hook, info)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		print(err.Error())
		return
	}
	contentTypeAddJSON(w.Header())
	json.NewEncoder(w).Encode(map[string]any{"id": id})
}

// group={groupid}&author={}&title={}&content={}
func (s *Server) pushToGroup(w http.ResponseWriter, r *ProcessedRequest) {
	fmt.Println("Endpoint Hit: pushToGroup")
	groupID := r.Params.Get("group")
	g, err := s.db.GetGroupByID(groupID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		print(err.Error())
		return
	}
	author := r.Params.Get("author")
	title := r.Params.Get("title")
	content := r.Params.Get("content")
	m := Message{Author: author, Title: title, Content: content}
	group := Group{g}
	resps, err := group.Push(&m)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		print(err.Error())
		return
	}
	succ := []string{}
	for _, resp := range resps {
		if resp.Err == nil {
			succ = append(succ, resp.Session.GetID())
		}
	}
	contentTypeAddJSON(w.Header())
	json.NewEncoder(w).Encode(succ)
}

// session={sessionid}&author={}&title={}&content={}
func (s *Server) pushToSession(w http.ResponseWriter, r *ProcessedRequest) {
	fmt.Println("Endpoint Hit: pushToSession")
	sessionID := r.Params.Get("session")
	se, err := s.db.GetSessionByID(sessionID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		print(err.Error())
		return
	}
	author := r.Params.Get("author")
	title := r.Params.Get("title")
	content := r.Params.Get("content")
	m := Message{Author: author, Title: title, Content: content}
	session := Session{se}
	_, err = session.Push(&m)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		print(err.Error())
		return
	}
}

// session={sessionid}
func (s *Server) checkSession(w http.ResponseWriter, r *ProcessedRequest) {
	fmt.Println("Endpoint Hit: checkSession")
	sessionID := r.Params.Get("session")
	se, err := s.db.GetSessionByID(sessionID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		print(err.Error())
		return
	}
	contentTypeAddJSON(w.Header())
	ret := map[string]any{"SessionInfo": se.GetInfo(), "PushHook": se.GetPushHook()}
	if g, err := se.GetGroup(); err != nil {
		ret["GroupID"] = g.GetID()
		ret["GroupInfo"] = g.GetInfo()
	}
	json.NewEncoder(w).Encode(ret)
}

// session={sessionid}
func (s *Server) hideSession(w http.ResponseWriter, r *ProcessedRequest) {
	fmt.Println("Endpoint Hit: hideSession")
	sessionID := r.Params.Get("session")
	se, err := s.db.GetSessionByID(sessionID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		print(err.Error())
		return
	}
	if err := se.Hide(); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		print(err.Error())
		return
	}
}
