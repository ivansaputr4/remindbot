package main

import (
	"fmt"
	"net/http"

	"github.com/ivansaputr4/remindbot/commands"
	"github.com/ivansaputr4/remindbot/config"
	"github.com/ivansaputr4/remindbot/handlers"
	"github.com/julienschmidt/httprouter"

	router "github.com/ivansaputr4/remindbot/router"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"

	"github.com/BurntSushi/toml"
	"github.com/justinas/alice"
)

func task(ac handlers.AppContext, chatId int64, text string) {
	ac.SendText(chatId, text)
}

func main() {
	var conf config.Config

	_, err := toml.DecodeFile("configs.toml", &conf)
	checkErr(err)

	fmt.Println(conf)
	db := initDB(conf.DB.Datapath)
	defer db.Close()

	ac := handlers.NewAppContext(db, conf, commands.NewCommandList())

	stack := alice.New()

	r := router.New()
	r.POST("/reminders", stack.ThenFunc(ac.CommandHandler))
	r.GET("/reminders", Healthz)

	fmt.Println("Server starting at port 1234.")
	http.ListenAndServe(":1234", r)
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

func Healthz(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, "ok")
}
