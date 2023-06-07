package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/turbitcat/tbcpusher/plugins/telegram/v2/wsgo"
	"github.com/turbitcat/tbcpusher/plugins/telegram/v2/wsgo/header"
)

type SessionInfo struct {
	ChatID   int64
	SenderID int64
}

func (s *SessionInfo) Recipient() string {
	return strconv.FormatInt(s.ChatID, 10)
}

func contentTypeIsJSON(h http.Header) bool {
	v, _ := header.ParseValueAndParams(h, "Content-Type")
	return v == "application/json"
}

func JoinGroup(u string, groupID string, callback string, data any) (string, error) {
	u, _ = url.JoinPath(u, "/session/create")
	ur, err := url.Parse(u)
	if err != nil {
		return "", fmt.Errorf("joinGroup parsing url: %v", err)
	}

	d := wsgo.H{"group": groupID, "hook": callback, "data": data}
	jsonData, err := json.Marshal(d)
	if err != nil {
		return "", fmt.Errorf("joinGroup marshal: %v", err)
	}
	resp, err := http.Post(ur.String(), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("joinGroup POST: %v", err)
	}
	if !contentTypeIsJSON(resp.Header) {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("joinGroup resp not json: %v", string(b))
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
	fmt.Printf("HideSession: GET %v\n", ur.String())
	resp, err := http.Get(ur.String())
	if err != nil {
		return fmt.Errorf("hideSession: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("hideSession: status %v", resp.StatusCode)
	}
	return nil
}

func GetSessionInfo(u string, sessionID string) (*SessionInfo, error) {
	u, _ = url.JoinPath(u, "/session/check")
	ur, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("checkSession parsing url: %v", err)
	}
	q := ur.Query()
	q.Add("session", sessionID)
	ur.RawQuery = q.Encode()
	fmt.Printf("JoinSession: GET %v\n", ur.String())
	resp, err := http.Get(ur.String())
	if err != nil {
		return nil, fmt.Errorf("checkSession GET: %v", err)
	}
	if !contentTypeIsJSON(resp.Header) {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("checkSession resp not json: %v", string(b))
	}
	var rs JSONSession
	if err := json.NewDecoder(resp.Body).Decode(&rs); err != nil {
		return nil, fmt.Errorf("checkSession parsing resp: %v", err)
	}
	info := rs.Data
	return &info, nil
}
