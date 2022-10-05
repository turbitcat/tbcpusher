package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/turbitcat/tpcpusher/v2/database"
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

func (s Session) Push(m *Message) (*http.Response, error) {
	url := s.GetPushHook()
	json_data, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("session push: %v", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(json_data))
	return resp, err
}

type pushResp struct {
	Session *Session
	Resp    *http.Response
	Err     error
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
