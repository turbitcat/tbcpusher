package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

func NewBot(token string) *telebot.Bot {
	pref := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}
	return b
}

func StartBotClient(b *telebot.Bot, adminIDs string, tbcpusherURL string, callbackURL string, callbackServer *CallbackServer) {
	adminOnly := b.Group()

	if adminIDs != "" {
		adminIDsStrings := strings.Split(adminIDs, ",")
		adminIDs := make([]int64, len(adminIDsStrings))
		for i, adminIDsString := range adminIDsStrings {
			n, err := fmt.Sscan(adminIDsString, &adminIDs[i])
			if err != nil {
				log.Fatal("error splite adminIDs: ", err)
				os.Exit(2)
			}
			if n != 1 {
				log.Fatal("error splite adminIDs")
				os.Exit(2)
			}
		}
		adminOnly.Use(middleware.Whitelist(adminIDs...))
	} else {
		adminOnly.Use(middleware.Whitelist())
	}

	b.Handle("/hello", func(c telebot.Context) error {
		c.Message()
		return c.Send(fmt.Sprintf("Hello, %v", c.Chat().ID))
	})

	adminOnly.Handle("/listremote", func(c telebot.Context) error {
		return c.Send("abcd")
	})

	b.Handle("/join", func(c telebot.Context) error {
		groupID := c.Message().Payload
		info := SessionInfo{ChatID: c.Chat().ID, SenderID: c.Sender().ID}
		sid, err := JoinGroup(tbcpusherURL, groupID, callbackURL, info)
		if err != nil {
			c.Send("err: " + err.Error())
			return err
		} else {
			return c.Send(fmt.Sprintf("Session id: %v", sid))
		}
	})

	b.Handle("/leave", func(c telebot.Context) error {
		sessionID := c.Message().Payload
		info, err := GetSessionInfo(tbcpusherURL, sessionID)
		if err != nil {
			c.Send("err: " + err.Error())
			return err
		}
		if info.SenderID == c.Sender().ID {
			if err = HideSession(tbcpusherURL, sessionID); err != nil {
				c.Send("err: " + err.Error())
				return err
			} else {
				return c.Send("left " + sessionID)
			}
		} else {
			return c.Send("You are not the owner of the session.")
		}
	})

	b.Handle("/info", func(c telebot.Context) error {
		pl := c.Message().Payload
		if pl == "on" {
			callbackServer.msgInfo = true
			return c.Send("Group id and Session id are going to show at the end of a message.")
		} else if pl == "off" {
			callbackServer.msgInfo = false
			return c.Send("Group id and Session id are hide now.")
		}
		return c.Send("/info [on|off]")
	})

	b.Start()
}
