package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/turbitcat/tbcpusher/v2/database"
	"github.com/turbitcat/tbcpusher/v2/scheduler"
	"github.com/turbitcat/tbcpusher/v2/wsgo"
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
	db        database.Database
	addr      string
	prefix    string
	router    *wsgo.ServerMux
	scheduler *scheduler.Scheduler
}

func NewServer(db database.Database) Server {
	s := wsgo.Default()
	s.Use(wsgo.ParseParamsJSON)
	sc := scheduler.NewDefult()
	return Server{db: db, addr: ":8000", router: s, scheduler: sc}
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

func (s *Server) Serve() error {
	r := s.router
	// get all groups
	r.Handle(s.prefix+"/group/all", func(c *wsgo.Context) {
		groups, err := s.db.GetAllGroups()
		if err != nil {
			c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}
		var ret []wsgo.H
		for _, g := range groups {
			group := Group{g}
			ret = append(ret, group.WsgoHWithSessions())
		}
		c.Json(http.StatusOK, ret)
	})
	// create a group
	// data={}
	r.Handle(s.prefix+"/group/create", func(c *wsgo.Context) {
		data, _ := c.Param("data")
		id, err := s.db.NewGroup(data)
		if err != nil {
			c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			return
		}
		c.Json(http.StatusOK, wsgo.H{"id": id})
	})
	// create a session
	// group={groupid}&hook={callbackurl}&data={}
	r.Handle(s.prefix+"/session/create", requireString("hook"), func(c *wsgo.Context) {
		ps := c.StringParams()
		gid, hook := ps["group"], ps["hook"]
		data, _ := c.Param("data")
		var sid string
		if gid != "" {
			g, err := s.db.GetGroupByID(gid)
			if err != nil {
				c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
				return
			}
			sid, err = g.NewSession(hook, data)
			if err != nil {
				c.String(http.StatusInternalServerError, http.StatusText(http.StatusBadRequest))
				return
			}
		} else {
			var err error
			sid, err = s.db.NewSession(hook, data)
			if err != nil {
				c.String(http.StatusInternalServerError, http.StatusText(http.StatusBadRequest))
				return
			}
		}
		c.Json(http.StatusOK, wsgo.H{"id": sid})
	})
	// push to group
	// group={groupid}&author={}&title={}&content={}
	r.Handle(s.prefix+"/group/push", requireString("group"), func(c *wsgo.Context) {
		gid, _ := c.StringParam("group")
		g, err := s.db.GetGroupByID(gid)
		if err != nil {
			c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			return
		}
		ps := c.StringParams()
		author, title, content := ps["author"], ps["title"], ps["content"]
		m := Message{author, title, content}
		if when, ok := c.StringParam("when"); ok {
			when_int, err := strconv.ParseInt(when, 10, 64)
			if err != nil {
				c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
				return
			}
			ti := time.Unix(when_int, 0)
			Group{g}.PushWhen(&m, ti, s.scheduler)
		} else {
			resps, err := Group{g}.Push(&m)
			if err != nil {
				c.String(http.StatusInternalServerError, http.StatusText(http.StatusBadRequest))
				return
			}
			succ := []string{}
			for _, resp := range resps {
				if resp.Err == nil {
					succ = append(succ, resp.Session.GetID())
				} else {
					c.LogIfLogging("push to group session %v: %v", resp.Session.GetID(), resp.Err)
				}
			}
			c.Json(http.StatusOK, succ)
		}
	})
	// push to session
	// session={sessionid}&author={}&title={}&content={}
	r.Handle(s.prefix+"/session/push", requireString("session"), func(c *wsgo.Context) {
		sid, _ := c.StringParam("session")
		session, err := s.db.GetSessionByID(sid)
		if err != nil {
			c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			return
		}
		ps := c.StringParams()
		author, title, content := ps["author"], ps["title"], ps["content"]
		m := Message{author, title, content}
		if when, ok := c.StringParam("when"); ok {
			when_int, err := strconv.ParseInt(when, 10, 64)
			if err != nil {
				c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
				return
			}
			ti := time.Unix(0, when_int*1000000)
			Session{session}.PushWhen(&m, ti, s.scheduler)
		} else {
			_, err = Session{session}.Push(&m)
			if err != nil {
				c.String(http.StatusNotAcceptable, http.StatusText(http.StatusNotAcceptable))
				return
			}
		}
	})
	// get session
	// session={sessionid}
	r.Handle(s.prefix+"/session/check", requireString("session"), func(c *wsgo.Context) {
		sid, _ := c.StringParam("session")
		session, err := s.db.GetSessionByID(sid)
		if err != nil {
			c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			return
		}
		ret := Session{session}.WsgoHWithGroup()
		c.Json(http.StatusOK, ret)
	})
	// hide session
	// session={sessionid}
	r.Handle(s.prefix+"/session/hide", requireString("session"), func(c *wsgo.Context) {
		sid, _ := c.StringParam("session")
		session, err := s.db.GetSessionByID(sid)
		if err != nil {
			c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			return
		}
		if err := session.Hide(); err != nil {
			c.String(http.StatusInternalServerError, http.StatusText(http.StatusBadRequest))
			return
		}
	})
	s.scheduler.Run()
	return r.Run(s.addr)
}
