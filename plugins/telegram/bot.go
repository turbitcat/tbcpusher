package main

import (
	"encoding/json"
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

func StartBotClient(b *telebot.Bot, adminIDs string, tbcpusherURL string, callBackURL string) {
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
		info := SessionInfo{ChatID: c.Chat().ID}
		_info, _ := json.Marshal(info)
		sid, err := JoinGroup(tbcpusherURL, groupID, callBackURL, string(_info))
		if err != nil {
			return c.Send("err: " + err.Error())
		} else {
			return c.Send(fmt.Sprintf("Session id: %v", sid))
		}
	})

	b.Start()
}
