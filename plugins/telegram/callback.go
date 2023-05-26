package main

import (
	"net/http"

	"github.com/turbitcat/tbcpusher/plugins/telegram/v2/wsgo"
	"gopkg.in/telebot.v3"
)

type CallbackServer struct {
	addr    string
	bot     *telebot.Bot
	prefix  string
	msgInfo bool
	route   *wsgo.ServerMux
}

func NewCallbackServer(bot *telebot.Bot) *CallbackServer {
	r := wsgo.Default()
	r.Use(wsgo.ParseParamsJSON)
	return &CallbackServer{bot: bot, addr: ":8001", msgInfo: true, route: r}
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
	s.route.POST(s.prefix+pathPush, s.receive)
	return s.route.Run(s.addr)
}

func (s *CallbackServer) CallbackPushURL() string {
	return s.prefix + pathPush
}

func (s *CallbackServer) receive(c *wsgo.Context) {
	var m JSONTBCPush
	if err := c.BindJSON(&m); err != nil {
		c.String(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		c.Log("Invalid body: %v\n", err)
		return
	}
	info := m.Session.Data
	msg := msgString(m.Message.Title, m.Message.Content, m.Message.Author)
	if s.msgInfo {
		msg = msg + "\n\nGroup: " + m.Session.GroupID + "\nSession: " + m.Session.Id
	}
	if _, err := s.bot.Send(&info, msg); err != nil {
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		c.Log("Error while send to tgbot: %v\n", err)
		return
	}
}

func msgString(title string, content string, author string) string {
	msg := ""
	mA := author
	mC := content
	mT := title
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
	return msg
}
