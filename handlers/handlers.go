package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	s "strings"
	"time"

	"github.com/ivansaputr4/remindbot/commands"
	"github.com/ivansaputr4/remindbot/config"
)

type Update struct {
	Id  int64   `json:"update_id"`
	Msg Message `json:"message"`
}

type Message struct {
	Id   int64  `json:"message_id"`
	Text string `json:"text"`
	Chat Chat   `json:"chat"`
}

type Chat struct {
	Id    int64  `json:"id"`
	Title string `json:"title"`
}

type AppContext struct {
	db   *sql.DB
	conf config.Config
	cmds commands.Commands
	loc  *time.Location
}

type Reminder struct {
	Id      int64     `sql:id`
	Content string    `sql:content`
	Created time.Time `sql:created`
	DueDt   time.Time `sql:due_dt`
	DueDay  string    `sql:due_day`
	ChatId  int64     `sql:chat_id`
}

func NewAppContext(db *sql.DB, conf config.Config, cmds commands.Commands) AppContext {
	us, _ := time.LoadLocation("Asia/Jakarta")
	return AppContext{db: db, conf: conf, cmds: cmds, loc: us}
}

func (ac *AppContext) CommandHandler(w http.ResponseWriter, r *http.Request) {
	var update Update

	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&update); err != nil {
		log.Println(err)
	} else {
		log.Println(update.Msg.Text)
	}

	cmd, txt, due_day, due_date := ac.cmds.Extract(update.Msg.Text)
	chatId := update.Msg.Chat.Id

	switch s.ToLower(cmd) {
	case "remind":
		ac.save(txt, due_day, due_date, chatId)
	case "check due":
		ac.CheckDue(chatId, false)
	case "list":
		ac.list(chatId)
	case "renum":
		ac.renum(chatId)
	case "clear":
		i, _ := strconv.Atoi(txt)
		ac.clear(i, chatId)
	case "clearall":
		ac.clearall(chatId)
	case "boti":
		ac.SendText(chatId, "Hellooo~~~~")
	}
}

func (ac *AppContext) save(txt string, due_day string, due_date time.Time, chatId int64) {
	now := time.Now().Format(time.RFC3339)

	_, err := ac.db.Exec(
		`INSERT INTO reminders(content, created, chat_id, due_dt, due_day) VALUES ($1, $2, $3, $4, $5)`,
		txt,
		now,
		chatId,
		due_date.Format(time.RFC3339),
		due_day)

	checkErr(err)
	ac.SendText(chatId, "Wait my reminders yaw!!!")
}

func (ac *AppContext) clear(id int, chatId int64) {
	_, err := ac.db.Exec(`DELETE FROM reminders WHERE chat_id=$1 AND id=$2`, chatId, id)
	checkErr(err)
	// "&#127881;"
	ac.SendText(chatId, "Bye!")
}

func (ac *AppContext) clearall(chatId int64) {
	_, err := ac.db.Exec(`DELETE FROM reminders WHERE chat_id=$1`, chatId)
	checkErr(err)
	ac.SendText(chatId, "Sayonaraaaaa!")
}

func (ac *AppContext) list(chatId int64) {
	rows, err := ac.db.Query(`SELECT id, content, due_dt, due_day FROM reminders WHERE chat_id=$1 and due_dt>=$2`, chatId, time.Now())
	checkErr(err)
	defer rows.Close()

	var arr []string
	var i int64
	var c string
	var due_date time.Time
	var due_day string

	listtype := [8]string{}
	listtype[0] = "monday"
	listtype[1] = "tuesday"
	listtype[2] = "wednesday"
	listtype[3] = "thursday"
	listtype[4] = "friday"
	listtype[5] = "saturday"
	listtype[6] = "sunday"

	for rows.Next() {
		_ = rows.Scan(&i, &c, &due_date, &due_day)
		line := "• " + c + " (`" + strconv.Itoa(int(i)) + "`)"

		if contains(listtype, due_day) {
			line = line + " - every " + due_day
		} else if due_day == "everyday" {
			line = line + " - " + due_day
		} else {
			line = line + " - "
		}

		if !due_date.IsZero() {
			line = line + " at " + due_date.In(ac.loc).Format("15:04:05")
		}

		arr = append(arr, line)
	}
	text := s.Join(arr, "\n")

	if len(arr) < 1 {
		text = "No current reminders :("
	}

	ac.SendText(chatId, text)
}

// This resets numbers for everyone!
func (ac *AppContext) renum(chatId int64) {
	rows, err := ac.db.Query(`SELECT content, due_dt, created, chat_id FROM reminders`)
	checkErr(err)
	defer rows.Close()

	var arr []Reminder
	var c string
	var dt time.Time
	var ct time.Time
	var chatid int64

	for rows.Next() {
		_ = rows.Scan(&c, &dt, &ct, &chatid)
		arr = append(arr, Reminder{Content: c, DueDt: dt, Created: ct, ChatId: chatid})
	}

	_, err = ac.db.Exec(`DELETE FROM reminders`)
	checkErr(err)

	_, err = ac.db.Exec(`DELETE FROM sqlite_sequence WHERE name='reminders';`)
	checkErr(err)

	for _, r := range arr {
		_, err := ac.db.Exec(`INSERT INTO reminders(content, due_dt, created, chat_id) VALUES ($1, $2, $3, $4)`, r.Content, r.DueDt, r.Created, r.ChatId)
		checkErr(err)
	}

	ac.list(chatId)
}

func (ac *AppContext) CheckDue(chatId int64, timedCheck bool) {
	timenow := time.Now().Add(time.Minute * 15)
	timepast := timenow.Add(time.Minute * 5)
	fmt.Println(timenow)
	fmt.Println("SELECT id, content, due_dt, due_day, chat_id FROM reminders WHERE due_dt>=" + timenow.Format(time.RFC3339) + " and due_dt<" + timepast.Format(time.RFC3339))
	rows, err := ac.db.Query(
		`SELECT id, content, due_dt, due_day, chat_id FROM reminders WHERE due_dt>=$1 and due_dt<$2`,
		timenow.Format(time.RFC3339),
		timepast.Format(time.RFC3339),
	)
	fmt.Println(err)
	fmt.Println(rows)

	checkErr(err)
	defer rows.Close()

	var id int64
	var content string
	var due_date time.Time
	var next_due_date time.Time
	var due_day string
	var chatid int64

	listtype := [8]string{}
	listtype[0] = "everyday"
	listtype[1] = "monday"
	listtype[2] = "tuesday"
	listtype[3] = "wednesday"
	listtype[4] = "thursday"
	listtype[5] = "friday"
	listtype[6] = "saturday"
	listtype[7] = "sunday"

	for rows.Next() {
		_ = rows.Scan(&id, &content, &due_date, &due_day, &chatid)
		line := "REMEMBER to " + content + " in 15 minutes."
		fmt.Println(line)
		ac.SendText(chatid, line)

		if contains(listtype, due_day) {
			if due_day == "everyday" {
				next_due_date = due_date.Add(time.Hour * 24 * 1)
			} else {
				next_due_date = due_date.Add(time.Hour * 24 * 7)
			}

			_, err := ac.db.Exec(
				`UPDATE reminders SET due_dt=$1 WHERE id=$2`,
				next_due_date.Format(time.RFC3339),
				id)

			checkErr(err)
		}
	}
}

func (ac *AppContext) SendText(chatId int64, text string) {
	link := "https://api.telegram.org/bot{apiKey}/sendMessage?chat_id={chatId}&text={text}&parse_mode=Markdown"
	// link = s.Replace(link, "{botId}", ac.conf.BOT.BotId, -1)
	link = s.Replace(link, "{apiKey}", ac.conf.BOT.ApiKey, -1)
	link = s.Replace(link, "{chatId}", strconv.FormatInt(chatId, 10), -1)
	link = s.Replace(link, "{text}", url.QueryEscape(text), -1)

	fmt.Println(link)

	_, _ = http.Get(link)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func contains(arr [8]string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func timeSinceLabel(d time.Time) string {
	var duration = time.Since(d)
	var durationNum int
	var unit string

	if int(duration.Hours()) == 0 {
		durationNum = int(duration.Minutes())
		unit = "min"
	} else if duration.Hours() < 24 {
		durationNum = int(duration.Hours())
		unit = "hour"
	} else {
		durationNum = int(duration.Hours()) / 24
		unit = "day"
	}

	if durationNum > 1 {
		unit = unit + "s"
	}

	return " `" + strconv.Itoa(int(durationNum)) + " " + unit + "`"
}
