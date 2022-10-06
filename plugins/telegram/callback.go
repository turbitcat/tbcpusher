package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"gopkg.in/telebot.v3"
)

type CallbackServer struct {
	addr   string
	bot    *telebot.Bot
	prefix string
}

func NewCallbackServer(bot *telebot.Bot) *CallbackServer {
	return &CallbackServer{bot: bot, addr: ":8001"}
}

func (s *CallbackServer) SetAddr(addr string) {
	s.addr = addr
}

func (s *CallbackServer) SetPrefix(p string) {
	if p != "" && p[0] != '/' {
		p = "/" + p
	}
	s.prefix = p
}

const pathPush = "/push"

func (s *CallbackServer) Serve() error {
	http.HandleFunc(s.prefix+pathPush, s.receive)
	return http.ListenAndServe(s.addr, nil)
}

func (s *CallbackServer) CallbackPushURL() string {
	return s.prefix + pathPush
}

type TBCPusherMessage struct {
	Author  string
	Title   string
	Content string
}

type JSONTBCPush struct {
	Msg         *TBCPusherMessage
	GroupID     string
	GroupInfo   string
	SessionID   string
	SessionInfo string
}

func (s *SessionInfo) Recipient() string {
	return strconv.FormatInt(s.ChatID, 10)
}

func (s *CallbackServer) receive(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: receive")
	var m JSONTBCPush
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		fmt.Println("bad request", err)
		return
	}
	jsonInfo := []byte(m.SessionInfo)
	var info SessionInfo
	if err := json.Unmarshal(jsonInfo, &info); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		fmt.Println("error parsing session info", err)
		return
	}
	var msg string
	mA := m.Msg.Author
	mC := m.Msg.Content
	mT := m.Msg.Title
	if mA == "" && mC == "" && mT == "" {
		msg = "Received an empty push."
	} else if mA != "" && mC == "" && mT == "" {
		msg = "Received an push from " + mA
	} else if mA == "" && mC != "" && mT == "" {
		msg = mC
	} else if mA != "" && mC != "" && mT == "" {
		msg = mA + "\n" + mC
	} else if mA == "" && mC == "" && mT != "" {
		msg = mT
	} else if mA != "" && mC == "" && mT != "" {
		msg = mA + "\n" + mT
	} else if mA == "" && mC != "" && mT != "" {
		msg = mT + "\n\n" + mC
	} else if mA != "" && mC != "" && mT != "" {
		msg = mA + "\n" + mT + "\n\n" + mC
	}
	if _, err := s.bot.Send(&info, msg); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		fmt.Println("error sending message to telegram", err)
		return
	}
}
