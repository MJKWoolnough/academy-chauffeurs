package main

import (
	"bytes"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/MJKWoolnough/ics"
)

func (c *Calls) calendar(w http.ResponseWriter, r *http.Request) {
	c.mu.Lock()
	cal, err := c.makeCalendar()
	c.mu.Unlock()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	err = ics.Encode(&buf, cal)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Length", strconv.Itoa(buf.Len()))
	w.Header().Add("Content-Type", "text/v-calendar; charset=utf-8")
	w.Header().Add("Content-Disposition", "attachment; filename=calendar.ics")
	buf.WriteTo(w)
}

func (c *Calls) makeCalendar() (*ics.Calendar, error) {
	var alarmTime int
	if err := c.statements[AlarmTime].QueryRow().Scan(&alarmTime); err != nil {
		return nil, err
	}
	var (
		a   ics.AlarmDisplay
		cal ics.Calendar
	)
	if alarmTime < 0 {
		a.Trigger.Duration = &ics.Duration{Negative: true, Minutes: uint(-alarmTime)}
	} else {
		a.Trigger.Duration = &ics.Duration{Minutes: uint(alarmTime)}
	}
	alarm := []ics.Alarm{{&a}}
	cal.ProductID = "Academy Chauffeurs 1.1"
	cal.Version = "2.0"
	n := now()
	rows, err := c.statements[CalendarData].Query((n - 3600*24*30*6) * 1000)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cal.Event = make([]ics.Event, 0, 1024)
	for rows.Next() {
		var (
			id, start, end, created, updated                        int64
			from, to, client, company, driverStr, phoneNumber, note string
			driver                                                  sql.NullString
		)
		err := rows.Scan(&id, &start, &end, &from, &to, &created, &updated, &note, &driver, &client, &company, &phoneNumber)
		if err != nil {
			return nil, err
		}
		from = strings.Replace(from, "\n", " ", -1)
		to = strings.Replace(to, "\n", " ", -1)
		if driver.Valid {
			driverStr = driver.String
		} else {
			driverStr = "Unassigned"
		}
		var ev ics.Event
		ev.UID = ics.PropUID(time.Unix(created, 0).In(time.UTC).Format("20060102T150405Z") + "-" + pad(strconv.FormatUint(uint64(id), 36)) + "@academy-chauffeurs.co.uk")
		ev.DateTimeStamp.Time = time.Unix(created, 0).In(time.UTC)
		ev.LastModified = &ics.PropLastModified{time.Unix(updated, 0).In(time.UTC)}
		startD := time.Unix(start/1000, start%1000).In(time.UTC)
		y, m, d := startD.Date()
		h, mi, s := startD.Clock()
		ev.DateTimeStart = &ics.PropDateTimeStart{DateTime: &ics.DateTime{time.Date(y, m, d, h, mi, s, 0, time.Local)}}
		var days, hours uint
		mins := uint(time.Unix(end/1000, end%1000).In(time.UTC).Sub(startD).Minutes())
		for mins > 60*24 {
			days++
			mins -= 60 * 24
		}
		for mins > 60 {
			hours++
			mins -= 60
		}
		ev.Duration = &ics.PropDuration{Days: days, Hours: hours, Minutes: mins}
		ev.Location = &ics.PropLocation{Text: ics.Text(from)}
		ev.Description = &ics.PropDescription{Text: ics.Text(driverStr + " - " + client + " (" + company + ") - " + from + " -> " + to + " - " + phoneNumber)}
		ev.Summary = &ics.PropSummary{Text: ics.Text(driverStr + " - " + client + " (" + company + ")")}
		ev.Alarm = alarm
		ev.Contact = []ics.PropContact{{Text: ics.Text(client + ", " + company + " - " + phoneNumber)}}
		ev.Comment = []ics.PropComment{{Text: ics.Text(notes)}}
		cal.Event = append(cal.Event, ev)
	}
	return &cal, nil
}

const padLength = 20

func pad(s string) string {
	t := bytes.Repeat([]byte{'0'}, padLength)
	copy(t[padLength-len(s):], s)
	return string(t)
}

// Errors

var ErrInvalidScheme = errors.New("invalid scheme")
