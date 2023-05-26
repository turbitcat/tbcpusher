package main

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/turbitcat/tbcpusher/plugins/telegram/v2/config"
)

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func main() {
	cfg := config.New()
	err := cfg.ReadAll(config.DefaultPath())
	if err != nil {
		println(err.Error())
	}
	if !fileExists("config.yml") {
		cfg.WriteFile(config.DefaultPath())
	}
	fmt.Printf("config: %+v\n", cfg)

	bot := NewBot(cfg.TelegramBot.Token)
	server := NewCallbackServer(bot)
	server.SetPrefix(cfg.Callback.Prefix)
	server.SetAddr(cfg.Callback.Address)
	go server.Serve()
	callbackurl, err := url.JoinPath(cfg.Callback.URLBase, server.CallbackPushURL())
	if err != nil {
		log.Fatalln("joinpath: " + err.Error())
	}
	fmt.Printf("Callbackurl: %v\n", callbackurl)
	StartBotClient(bot, cfg.TelegramBot.AdminIDs, cfg.TBCPusher.URL, callbackurl, server)
}
