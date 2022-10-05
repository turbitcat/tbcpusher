package main

import (
	"fmt"
	"os"

	"github.com/turbitcat/tpcpusher/v2/api"
	"github.com/turbitcat/tpcpusher/v2/config"
	"github.com/turbitcat/tpcpusher/v2/database"
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
	db, err := database.NewMongo(cfg.Mongo.AtlasURI, cfg.Mongo.Database)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	server := api.NewServer(db)
	server.SetAddr(cfg.Api.Address)
	server.SetPrefix(cfg.Api.Prefix)
	server.Serve()
}
