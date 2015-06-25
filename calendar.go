package main

import (
	"bytes"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
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
	err = ics.NewEncoder(&buf).Encode(cal)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Length", strconv.Itoa(buf.Len()))
	buf.WriteTo(w)
}

func (c *Calls) makeCalendar() (*ics.Calendar, error) {
	var alarmTime int
	if err := c.statements[AlarmTime].QueryRow().Scan(&alarmTime); err != nil {
		return nil, err
	}
	var cal ics.Calendar
	cal.ProductID = "CALExport 0.01"
	n := now()
	rows, err := c.statements[CalendarData].Query((n-3600*24*31)*1000, (n+3600*24*365)*1000)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cal.Events = make([]ics.Event, 0, 1024)
	for rows.Next() {
		var (
			id, start, end, created, updated     int64
			from, to, client, company, driverStr string
			driver                               sql.NullString
		)
		err := rows.Scan(&id, &start, &end, &from, &to, &created, &updated, &driver, &client, &company)
		if err != nil {
			return nil, err
		}
		if driver.Valid {
			driverStr = driver.String
		} else {
			driverStr = "Unassigned"
		}
		ev := ics.NewEvent()
		ev.UID = time.Unix(created, 0).In(time.UTC).Format("20060102T150405Z") + "-" + pad(strconv.FormatUint(uint64(id), 36)) + "@academy-chauffeurs.co.uk"
		ev.Created = time.Unix(created, 0).In(time.UTC)
		ev.LastModified = time.Unix(updated, 0).In(time.UTC)
		ev.Start.Time = time.Unix(start/1000, start%1000)
		ev.Duration.Duration = time.Unix(end/1000, end%1000).Sub(ev.Start.Time)
		ev.Location.String = from
		ev.Description.String = driverStr + " - " + client + " (" + company + ") - " + from + " -> " + to
		ev.Summary.String = driverStr + " - " + client + " (" + company + ")"
		var a ics.DisplayAlarm
		a.Trigger.Duration = time.Minute * alarmTime
		ev.Alarms = []ics.Alarm{a}
		cal.Events = append(cal.Events, ev)
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
