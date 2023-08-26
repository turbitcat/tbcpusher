package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/turbitcat/tbcpusher/v2/database"
	"github.com/turbitcat/tbcpusher/v2/scheduler"
	"github.com/turbitcat/tbcpusher/v2/wsgo"
)

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
	list := database.NewEntryList(db, ScheduleGetter, JobGetter)
	sc.SetEntries(list)
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
	// r.Handle(s.prefix+"/group/all", func(c *wsgo.Context) {
	// 	groups, err := s.db.GetAllGroups()
	// 	if err != nil {
	// 		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	// 		c.Log("%v", err)
	// 		return
	// 	}
	// 	var ret []wsgo.H
	// 	for _, g := range groups {
	// 		group := Group{g}
	// 		ret = append(ret, group.WsgoHWithSessions())
	// 	}
	// 	c.Json(http.StatusOK, ret)
	// })
	// create a group
	// data={}
	r.Handle(s.prefix+"/doc", func(c *wsgo.Context) {
		c.FormatedJson(http.StatusOK, Docs)
	})
	r.Handle(s.prefix+"/group/create", func(c *wsgo.Context) {
		data, _ := c.Param("data")
		id, err := s.db.NewGroup(data)
		if err != nil {
			c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			c.Log("Bad Request: %v", err)
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
				c.Log("Bad Request: %v", err)
				return
			}
			sid, err = g.NewSession(hook, data)
			if err != nil {
				c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
				c.Log("NewSession: %v", err)
				return
			}
		} else {
			var err error
			sid, err = s.db.NewSession(hook, data)
			if err != nil {
				c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
				c.Log("NewSession: %v", err)
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
			c.Log("Bad Request: %v", err)
			return
		}
		ps := c.StringParams()
		author, title, content := ps["author"], ps["title"], ps["content"]
		m := Message{author, title, content}
		if when, ok := c.StringParam("when"); ok {
			when_int, err := strconv.ParseInt(when, 10, 64)
			if err != nil {
				c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
				c.Log("Bad Request: %v", err)
				return
			}
			ti := time.Unix(0, when_int*1000000)
			Group{g}.PushWhen(&m, ti, s.scheduler)
		} else {
			resps, err := Group{g}.Push(&m)
			if err != nil {
				c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
				c.Log("push to group %v: %v", gid, err)
				return
			}
			succ := []string{}
			for _, resp := range resps {
				if resp.Err == nil {
					succ = append(succ, resp.Session.GetID())
				} else {
					c.Log("push to group session %v: %v", resp.Session.GetID(), resp.Err)
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
			c.Log("Bad Request: %v", err)
			return
		}
		ps := c.StringParams()
		author, title, content := ps["author"], ps["title"], ps["content"]
		m := Message{author, title, content}
		if when, ok := c.StringParam("when"); ok {
			when_int, err := strconv.ParseInt(when, 10, 64)
			if err != nil {
				c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
				c.Log("Bad Request: %v", err)
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
			c.Log("Bad Request: %v", err)
			return
		}
		ret := Session{session}.WsgoHWithGroup()
		c.Json(http.StatusOK, ret)
	})
	// set session data
	// session={sessionid}
	r.Handle(s.prefix+"/session/setdata", requireString("session"), func(c *wsgo.Context) {
		sid, _ := c.StringParam("session")
		data, _ := c.Param("data")
		session, err := s.db.GetSessionByID(sid)
		if err != nil {
			c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			c.Log("Bad Request: %v", err)
			return
		}
		if err := session.SetData(data); err != nil {
			c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			c.Log("session %v set data %v: %v", sid, data, err)
			return
		}
	})
	// set group data
	// group={groupid}
	r.Handle(s.prefix+"/group/setdata", requireString("group"), func(c *wsgo.Context) {
		gid, _ := c.StringParam("group")
		data, _ := c.Param("data")
		group, err := s.db.GetGroupByID(gid)
		if err != nil {
			c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			c.Log("Bad Request: %v", err)
			return
		}
		if err := group.SetData(data); err != nil {
			c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			c.Log("group %v set data %v: %v", gid, data, err)
			return
		}
	})
	// hide session
	// session={sessionid}
	r.Handle(s.prefix+"/session/hide", requireString("session"), func(c *wsgo.Context) {
		sid, _ := c.StringParam("session")
		session, err := s.db.GetSessionByID(sid)
		if err != nil {
			c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
			c.Log("Bad Request: %v", err)
			return
		}
		if err := session.Hide(); err != nil {
			c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			c.Log("hide session %v: %v", sid, err)
			return
		}
	})
	s.scheduler.Run()
	return r.Run(s.addr)
}
