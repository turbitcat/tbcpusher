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
	// fmt.Println(db.NewGroup("abab"))
	// gs, _ := db.GetAllGroups()
	// g := gs[0]
	// fmt.Println(g)
	// fmt.Println(g.GetSessions())
	// fmt.Println(g.NewSession("ccccc"))
	// fmt.Println(g.GetSessions())
	api.Serve()
}
