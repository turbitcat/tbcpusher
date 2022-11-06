package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/turbitcat/tbcpusher/v2/database"
	"github.com/turbitcat/tbcpusher/v2/scheduler"
	"github.com/turbitcat/tbcpusher/v2/wsgo"
)

type Message struct {
	Author  string
	Title   string
	Content string
}

type Group struct {
	database.Group
}

type Session struct {
	database.Session
}

func (s Session) WsgoH() wsgo.H {
	return wsgo.H{"id": s.GetID(), "data": s.GetData(), "hook": s.GetPushHook(), "groupID": s.GetGroupID()}
}

func (s Session) WsgoHWithGroup() wsgo.H {
	r := s.WsgoH()
	g, err := s.GetGroup()
	if err == nil {
		r["group"] = Group{g}.WsgoH()
	}
	return r
}

func (s Session) Push(m *Message) (*http.Response, error) {
	url := s.GetPushHook()
	data := wsgo.H{"session": s.WsgoHWithGroup(), "message": m}
	json_data, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("session push: %v", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		err = fmt.Errorf("session push: %v", err)
	}
	return resp, err
}

type pushResp struct {
	Session *Session
	Resp    *http.Response
	Err     error
}

func (g Group) WsgoH() wsgo.H {
	return wsgo.H{"id": g.GetID(), "data": g.GetData()}
}

func (g Group) WsgoHWithSessions() wsgo.H {
	r := g.WsgoH()
	ss := []wsgo.H{}
	gs, err := g.GetSessions()
	if err == nil {
		for _, s := range gs {
			ss = append(ss, Session{s}.WsgoH())
		}
	}
	r["sessions"] = ss
	return r
}

func (g Group) Push(m *Message) ([]pushResp, error) {
	sessions, err := g.GetSessions()
	if err != nil {
		return nil, err
	}
	l := []pushResp{}
	for _, s := range sessions {
		res, err := Session{s}.Push(m)
		l = append(l, pushResp{&Session{s}, res, err})
	}
	return l, nil
}

func (s Session) PushWhen(m *Message, t time.Time, sc *scheduler.Scheduler) error {
	url := s.GetPushHook()
	data := wsgo.H{"session": s.WsgoHWithGroup(), "message": m}
	json_data, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("session pushWhen: %v", err)
	}
	f := func() { http.Post(url, "application/json", bytes.NewBuffer(json_data)) }
	ti := &scheduler.OneTimeSchedule{T: t}
	sc.AddFunc(f, ti)
	return nil
}

func (g Group) PushWhen(m *Message, t time.Time, sc *scheduler.Scheduler) error {
	sessions, err := g.GetSessions()
	if err != nil {
		return fmt.Errorf("group pushWhen: %v", err)
	}
	for _, s := range sessions {
		Session{s}.PushWhen(m, t, sc)
	}
	return nil
}
