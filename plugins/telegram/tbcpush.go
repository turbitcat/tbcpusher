package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type SessionInfo struct {
	ChatID int64
}

func JoinGroup(u string, groupId string, callback string, info string) (string, error) {
	u, _ = url.JoinPath(u, "/session/create")
	ur, err := url.Parse(u)
	if err != nil {
		return "", fmt.Errorf("joinGroup parsing url: %v", err)
	}
	q := ur.Query()
	q.Add("group", groupId)
	q.Add("hook", callback)
	q.Add("info", info)
	ur.RawQuery = q.Encode()
	fmt.Println(ur.String())
	resp, err := http.Get(ur.String())
	if err != nil {
		return "", fmt.Errorf("joinGroup err: %v", err)
	}
	var rs struct{ Id string }
	if err := json.NewDecoder(resp.Body).Decode(&rs); err != nil {
		return "", fmt.Errorf("joinGroup parsing resp: %v", err)
	}
	return rs.Id, nil
}
