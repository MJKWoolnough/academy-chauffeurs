package main

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/MJKWoolnough/ics"
	"github.com/jlaffaye/ftp"
)

var (
	calChan chan struct{}
)

func (c *Calls) uploader() {
	t := time.NewTicker(10 * time.Minute)
	for {
		select {
		case <-t.C:
			var modified byte
			c.mu.Lock()
			err := c.statements[IsModified].QueryRow().Scan(&modified)
			if err == nil && modified == 1 {
				err = c.uploadCalendar()
			}
			c.mu.Unlock()
			if err != nil {
				log.Println(err)
			}
		case <-calChan:
			t.Stop()
			return
		}
	}
}

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

var (
	calMut                                          sync.RWMutex
	calendarUsername, calendarPassword, calendarURL string
)

func (c *Calls) uploadCalendar() error {
	calMut.RLock()
	calUsername, calPassword, calURL := calendarUsername, calendarPassword, calendarURL
	calMut.RUnlock()
	cal, err := c.makeCalendar()
	if err != nil {
		return err
	}
	uri, err := url.Parse(calURL)
	if err != nil {
		return err
	}
	conn, err := ftp.Dial(uri.Host)
	if err != nil {
		return err
	}
	err = conn.Login(calUsername, calPassword)
	if err != nil {
		return err
	}
	pr, pw := io.Pipe()
	defer pr.Close()
	go func() {
		defer pw.Close()
		ics.NewEncoder(pw).Encode(cal)
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
	return nil
}

func (c *Calls) makeCalendar() (*ics.Calendar, error) {
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
			id, start, end, created, updated  int64
			from, to, driver, client, company string
		)
		err := rows.Scan(&id, &start, &end, &from, &to, &created, &updated, &driver, &client, &company)
		if err != nil {
			return nil, err
		}
		ev := ics.NewEvent()
		ev.UID = time.Unix(created, 0).In(time.UTC).Format("20060102T150405Z") + "-" + pad(strconv.FormatUint(uint64(id), 36)) + "@academy-chauffeurs.co.uk"
		ev.Created = time.Unix(created, 0).In(time.UTC)
		ev.LastModified = time.Unix(updated, 0).In(time.UTC)
		ev.Start.Time = time.Unix(start/1000, start%1000).In(time.Local)
		ev.Duration.Duration = time.Unix(end/1000, end%1000).Sub(ev.Start.Time)
		ev.Location.String = from
		ev.Description.String = driver + " - " + client + " (" + company + ") - " + from + " -> " + to
		ev.Summary.String = driver + " - " + client + " (" + company + ")"
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
