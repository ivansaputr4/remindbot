package commands

import (
	"fmt"
	"regexp"
	s "strings"
	"time"

	"github.com/jinzhu/now"
)

type Commands struct {
	rmt  *regexp.Regexp
	cd   *regexp.Regexp
	r    *regexp.Regexp
	l    *regexp.Regexp
	rn   *regexp.Regexp
	c    *regexp.Regexp
	cl   *regexp.Regexp
	boti *regexp.Regexp
}

func NewCommandList() Commands {
	now.TimeFormats = append(now.TimeFormats, "2Jan 15:04 2006")
	now.TimeFormats = append(now.TimeFormats, "2Jan 3:04pm 2006")
	now.TimeFormats = append(now.TimeFormats, "2Jan 3pm 2006")

	now.TimeFormats = append(now.TimeFormats, "2Jan 2006 15:04")
	now.TimeFormats = append(now.TimeFormats, "2Jan 2006 3:04pm")
	now.TimeFormats = append(now.TimeFormats, "2Jan 2006 3pm")

	now.TimeFormats = append(now.TimeFormats, "2Jan 15:04")
	now.TimeFormats = append(now.TimeFormats, "2Jan 3:04pm")
	now.TimeFormats = append(now.TimeFormats, "2Jan 3pm")

	now.TimeFormats = append(now.TimeFormats, "2Jan")

	return Commands{
		rmt:  compileRegexp(`(?im)(remind){1}(?: me to)? ([^:\r\n]*)(?::?)([^:\r\n]*)(?::?)(.*)$`),
		cd:   compileRegexp(`(?im)(check due)$`),
		l:    compileRegexp(`(?im)(list)$`),
		c:    compileRegexp(`(?im)(clear) (\d+)$`),
		rn:   compileRegexp(`(?im)(renum)$`),
		cl:   compileRegexp(`(?im)(clearall)$`),
		boti: compileRegexp(`(?im)(boti)(?:!|~)?$`),
	}
}

func compileRegexp(s string) *regexp.Regexp {
	r, _ := regexp.Compile(s)
	return r
}

func (c *Commands) Extract(t string) (string, string, string, time.Time) {
	var a []string
	var r1, r2, r3, r4 = "", "", "", ""
	fmt.Println("-----")
	fmt.Println(t)
	id, _ := time.LoadLocation("Asia/Jakarta")
	// utc, _ := time.LoadLocation("UTC")

	a = c.rmt.FindStringSubmatch(t)
	if len(a) == 5 {
		r1, r2, r3, r4 = a[1], a[2], a[3], a[4]
	}

	a = c.cd.FindStringSubmatch(t)
	if len(a) == 2 {
		r1 = a[1]
	}

	a = c.l.FindStringSubmatch(t)
	if len(a) == 2 {
		r1 = a[1]
	}

	a = c.c.FindStringSubmatch(t)
	if len(a) == 3 {
		r1, r2 = a[1], a[2]
	}

	a = c.rn.FindStringSubmatch(t)
	if len(a) == 2 {
		r1 = a[1]
	}

	a = c.cl.FindStringSubmatch(t)
	if len(a) == 2 {
		r1 = a[1]
	}

	a = c.boti.FindStringSubmatch(t)
	if len(a) == 2 {
		r1 = a[1]
	}

	r1 = s.ToLower(s.TrimSpace(r1))
	r2 = s.ToLower(s.TrimSpace(r2))
	r3 = s.ToLower(s.TrimSpace(r3))
	r4 = s.ToLower(s.TrimSpace(r4))
	fmt.Println("---")
	fmt.Println(r1)
	fmt.Println(r2)
	fmt.Println(r3)
	fmt.Println(r4)
	fmt.Println("---")

	tmrRegex := regexp.MustCompile("(?im)^tomorrow|tmr|tml")
	todayRegex := regexp.MustCompile("(?im)today")
	weekdayRegex := regexp.MustCompile("(?im)everyday|monday|tuesday|wednesday|thursday|friday|saturday|sunday") //weekday or everyday
	if tmrRegex.MatchString(r3) {
		r4 = time.Now().AddDate(0, 0, 1).Format("2Jan") + " " + r4
		r3 = "default"
	} else if todayRegex.MatchString(r3) {
		r4 = time.Now().Format("2Jan") + " " + r4
		r3 = "default"
	} else if weekdayRegex.MatchString(r3) {
		// do nothing
	} else {
		r4 = r3 + " " + r4
		r3 = "default"
	}

	fmt.Println(r3)
	fmt.Println(r4)
	due_date, err := now.Parse(r4)
	fmt.Println(due_date.String())
	var r4t time.Time
	if err == nil {
		// remind at 15 min before
		due_date = time.Date(due_date.Year(), due_date.Month(), due_date.Day(), due_date.Hour(), due_date.Minute(), 0, 0, id)
		r4t = due_date.Local()
	} else {
		r4t = time.Time{}
	}

	fmt.Println("---")
	fmt.Println(r1)
	fmt.Println(r2)
	fmt.Println(r3)
	fmt.Println(r4t)
	fmt.Println("-----------")
	return r1, r2, r3, r4t
}
