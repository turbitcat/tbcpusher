package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/turbitcat/tpcpusher/v2/database"
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
	db     database.Database
	addr   string
	prefix string
}

const pathCreateGroup = "/group/create"
const pathPushToGroup = "/group/push"
const pathCreateSession = "/session/create"

func NewServer(db database.Database) Server {
	return Server{db: db, addr: ":8000"}
}

func (s *Server) SetAddr(addr string) {
	s.addr = addr
}

func (s *Server) SetPrefix(p string) {
	if p == "" {
		return
	}
	if p[0] != '/' {
		p = "/" + p
	}
	s.prefix = p
}

func (s *Server) Serve() error {
	http.HandleFunc(s.prefix+pathCreateGroup, s.createGroup)
	http.HandleFunc(s.prefix+pathPushToGroup, s.pushToGroup)
	http.HandleFunc(s.prefix+pathCreateSession, s.createSession)
	return http.ListenAndServe(s.addr, nil)
}

// info={}
func (s *Server) createGroup(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: createGroup")
	info := getParamStringFromURLAndBody(r, "info")
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
func (s *Server) createSession(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: createSession")
	groupID := getParamStringFromURLAndBody(r, "group")
	g, err := s.db.GetGroupByID(groupID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		print(err.Error())
		return
	}
	hook := getParamStringFromURLAndBody(r, "hook")
	if hook == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		print("hook url is empty")
		return
	}
	info := getParamStringFromURLAndBody(r, "info")
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
func (s *Server) pushToGroup(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: pushToGroup")
	groupID := getParamStringFromURLAndBody(r, "group")
	g, err := s.db.GetGroupByID(groupID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		print(err.Error())
		return
	}
	author := getParamStringFromURLAndBody(r, "author")
	title := getParamStringFromURLAndBody(r, "title")
	content := getParamStringFromURLAndBody(r, "content")
	m := Message{Author: author, Title: title, Content: content}
	group := Group{g}
	resps, err := group.Push(&m)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		print(err.Error())
		return
	}
	errs := []string{}
	for _, resp := range resps {
		if resp.Err == nil {
			errs = append(errs, resp.Session.GetID())
		}
	}
	contentTypeAddJSON(w.Header())
	json.NewEncoder(w).Encode(errs)
}
