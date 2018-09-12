package main

import (
	"fmt"

	"github.com/ivansaputr4/remindbot/commands"
	"github.com/ivansaputr4/remindbot/config"
	"github.com/ivansaputr4/remindbot/handlers"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"

	"github.com/BurntSushi/toml"
	"github.com/jasonlvhit/gocron"
)

func main() {
	var conf config.Config

	_, err := toml.DecodeFile("configs.toml", &conf)
	checkErr(err)

	fmt.Println(conf)
	db := initDB(conf.DB.Datapath)
	defer db.Close()

	ac := handlers.NewAppContext(db, conf, commands.NewCommandList())
	gocron.Every(5).Minutes().Do(ac.CheckDue, conf.BOT.MainChatId, true)
	fmt.Println("Starting timer")
	<-gocron.Start()
}

func initDB(datapath string) *sql.DB {
	db, err := sql.Open("sqlite3", datapath+"/reminders.db")
	checkErr(err)
	return db
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
