package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/golang/gddo/httputil/header"
)

type SessionInfo struct {
	ChatID   int64
	SenderID int64
}

func contentTypeIsJSON(h http.Header) bool {
	v, _ := header.ParseValueAndParams(h, "Content-Type")
	return v == "application/json"
}

func JoinGroup(u string, groupID string, callback string, info string) (string, error) {
	u, _ = url.JoinPath(u, "/session/create")
	ur, err := url.Parse(u)
	if err != nil {
		return "", fmt.Errorf("joinGroup parsing url: %v", err)
	}
	q := ur.Query()
	q.Add("group", groupID)
	q.Add("hook", callback)
	q.Add("info", info)
	ur.RawQuery = q.Encode()
	fmt.Printf("JoinGroup: GET %v", ur.String())
	resp, err := http.Get(ur.String())
	if err != nil {
		return "", fmt.Errorf("joinGroup GET: %v", err)
	}
	if !contentTypeIsJSON(resp.Header) {
		b, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("joinGroup resp not json: %v", b)
	}
	var rs struct{ Id string }
	if err := json.NewDecoder(resp.Body).Decode(&rs); err != nil {
		return "", fmt.Errorf("joinGroup parsing resp: %v", err)
	}
	return rs.Id, nil
}

func HideSession(u string, sessionID string) error {
	u, _ = url.JoinPath(u, "/session/hide")
	ur, err := url.Parse(u)
	if err != nil {
		return fmt.Errorf("checkSession parsing url: %v", err)
	}
	q := ur.Query()
	q.Add("session", sessionID)
	ur.RawQuery = q.Encode()
	fmt.Printf("HideSession: GET %v", ur.String())
	resp, err := http.Get(ur.String())
	if err != nil {
		return fmt.Errorf("hideSession: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("hideSession: status %v", resp.StatusCode)
	}
	return nil
}

func CheckSession(u string, sessionID string) (*SessionInfo, error) {
	u, _ = url.JoinPath(u, "/session/check")
	ur, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("checkSession parsing url: %v", err)
	}
	q := ur.Query()
	q.Add("session", sessionID)
	ur.RawQuery = q.Encode()
	fmt.Printf("JoinSession: GET %v", ur.String())
	resp, err := http.Get(ur.String())
	if err != nil {
		return nil, fmt.Errorf("checkSession GET: %v", err)
	}
	if !contentTypeIsJSON(resp.Header) {
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("checkSession resp not json: %v", string(b))
	}
	var rs struct{ SessionInfo string }
	if err := json.NewDecoder(resp.Body).Decode(&rs); err != nil {
		return nil, fmt.Errorf("checkSession parsing resp: %v", err)
	}
	var info SessionInfo
	if err := json.Unmarshal([]byte(rs.SessionInfo), &info); err != nil {
		return nil, fmt.Errorf("checkSession parsing session info: %v", err)
	}
	return &info, nil
}
