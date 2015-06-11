package main

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/MJKWoolnough/ics"
	"github.com/jlaffaye/ftp"
)

func (c *Calls) calendar(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	cal, err := c.makeCalendar()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	ics.NewEncoder(&buf).Encode(cal)
	w.Header().Add("Content-Length", strconv.Itoa(buf.Len()))
	buf.WriteTo(w)
}

var (
	calMut                                          sync.RWMutex
	calendarUsername, calendarPassword, calendarURL string
	uploadCalendar                                  bool
)

func (c *Calls) uploadCalendar() error {
	calMut.RLock()
	defer calMut.RUnlock()
	if !uploadCalendar {
		return nil
	}
	cal, err := c.makeCalendar()
	if err != nil {
		return err
	}
	uri, err := url.Parse(calendarURL)
	if err != nil {
		return err
	}
	conn, err := ftp.Dial(uri.Host)
	if err != nil {
		return err
	}
	err = conn.Login(calendarUsername, calendarPassword)
	if err != nil {
		return err
	}
	pr, pw := io.Pipe()
	defer pr.Close()
	go func() {
		ics.NewEncoder(pw).Encode(cal)
		pw.Close()
	}()
	return conn.Stor(uri.Path, pr)
}

func checkUpload(upload bool, username, password, u string) error {
	if upload {
		uri, err := url.Parse(u)
		if err != nil {
			return err
		}
		if uri.Scheme != "ftp" {
			return ErrInvalidScheme
		}
		conn, err := ftp.Dial(uri.Host)
		if err != nil {
			return err
		}
		err = conn.Login(username, password)
		if err != nil {
			return err
		}
	}
	calMut.Lock()
	defer calMut.Unlock()
	calendarUsername = username
	calendarPassword = password
	calendarURL = u
	uploadCalendar = upload
	return nil
}

func (c *Calls) makeCalendar() (*ics.Calendar, error) {
	var cal ics.Calendar
	cal.ProductID = "CALExport 0.01"
	n := now()
	rows, err := c.statements[CalendarData].Query(n-3600*24*31, n+3600*24*365)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cal.Events = make([]ics.Event, 0, 1024)
	for rows.Next() {
		var (
			start, end, created, updated      int64
			from, to, driver, client, company string
		)
		err := rows.Scan(&start, &end, &from, &to, &created, &updated, &driver, &client, &company)
		if err != nil {
			return nil, err
		}
		ev := ics.NewEvent()
		ev.Created = time.Unix(created, 0)
		ev.LastModified = time.Unix(updated, 0)
		ev.Start.Time = time.Unix(start/1000, 0).In(time.Local)
		ev.Duration.Duration = time.Unix(end/1000, 0).Sub(time.Unix(start/1000, 0))
		ev.Location.String = from
		ev.Description.String = driver + " - " + client + " (" + company + ") - " + from + " -> " + to
		cal.Events = append(cal.Events, ev)
	}
	return &cal, nil
}

// Errors

var ErrInvalidScheme = errors.New("invalid scheme")
