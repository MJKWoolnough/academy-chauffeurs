package main

import (
	"net/http"
	"time"

	"github.com/MJKWoolnough/form"
	"github.com/MJKWoolnough/store"
)

type Event struct {
	ID       int
	Start    time.Time
	End      time.Time
	From, To string
	Pickup   time.Time
	Waiting  int64
	Parking  int
	Miles    int
	ClientId int
	Driver
}

func (e *Event) Get() store.TypeMap {
	return store.TypeMap{
		"id":                 &e.ID,
		"start":              &e.Start,
		"end":                &e.End,
		"from":               &e.From,
		"to":                 &e.To,
		"pickupTime":         &e.Pickup,
		"waitingTime":        &e.Waiting,
		"parkingCosts":       &e.Parking,
		"milesDriver":        &e.Miles,
		"clientId":           &e.ClientId,
		"driverId":           &e.Driver.ID,
		"driverName":         &e.Driver.Name,
		"driverRegistration": &e.Driver.Registration,
		"driverPhone":        &e.Driver.Phone,
	}
}

func (e *Event) ParserList() form.ParserList {
	return form.ParserList{
		"id":                 form.Int{&e.ID},
		"start":              form.UnixTime{&e.Start},
		"end":                form.UnixTime{&e.End},
		"from":               form.String{&e.From},
		"to":                 form.String{&e.To},
		"pickupTime":         form.Time{&e.Pickup},
		"waitingTime":        form.Int64{&e.Waiting},
		"parkingCosts":       form.Int{&e.Parking},
		"milesDriver":        form.Int{&e.Miles},
		"clientId":           form.Int{&e.ClientId},
		"driverId":           form.Int{&e.Driver.ID},
		"driverName":         form.String{&e.Driver.Name},
		"driverRegistration": form.String{&e.Driver.Registration},
		"driverPhone":        form.String{&e.Driver.Phone},
	}
}

func (Event) Key() string {
	return "id"
}

func (Event) TableName() string {
	return "events"
}

func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	var t time.Time
	r.ParseForm()
	err := form.ParseValue("date", form.TimeFormat{&t, dateFormat}, r.Form)
	if err != nil || r.Form.Get("date") == "" {
		t = time.Now()
	}
	s.eventList(w, r, t, ModeNormal, nil)
}

type EventErrors struct {
	Event
	FromError, ToError string
}

func (s *Server) addEvent(w http.ResponseWriter, r *http.Request) {
	var e EventErrors
	r.ParseForm()
	form.Parse(&e, r.PostForm)
	s.addUpdateEvent(w, r, &e, 0)
}

func (s *Server) addUpdateEvent(w http.ResponseWriter, r *http.Request, e *EventErrors, forcePage int) {
	date := time.Now()
	form.ParseValue("date", form.TimeFormat{&date, dateFormat}, r.PostForm)
	if e.Start.IsZero() {
		forcePage = 1
		e.Start = time.Now()
	}
	if forcePage == 1 {
		e.Start = normaliseDay(e.Start)
		s.eventList(w, r, date, ModeStart, &e.Event)
		return
	}
	if e.End.IsZero() || !e.End.After(e.Start) || forcePage == 2 {
		e.End = normaliseDay(e.Start)
		s.eventList(w, r, date, ModeEnd, &e.Event)
		return
	}
	if r.PostForm.Get("submit") != "" {
		good := true
		// Check values for errors - set errors
		if good {
			// add to/update store
			http.Redirect(w, r, "events?date="+e.Start.Format(dateFormat), http.StatusFound)
			return
		}
	}
	s.pages.ExecuteTemplate(w, "addUpdateEvent.html", e)
}

func (s *Server) updateEvent(w http.ResponseWriter, r *http.Request) {
	var (
		e         EventErrors
		forcePage int
	)
	r.ParseForm()
	form.ParseValue("id", form.Int{&e.ID})
	if e.ID == 0 {
		http.Redirect(w, r, "events", http.StatusFound)
		return
	}
	err := s.db.Get(e)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if e.ID == 0 {
		http.Redirect(w, r, "events", http.StatusFound)
		return
	}
	form.Parse(&e, r.PostForm)
	form.ParseValue("force", form.Int{&forcePage}, r.PostForm)
	s.addUpdateEvent(w, r, &e, forcePage)
}

func (s *Server) removeEvent(w http.ResponseWriter, r *http.Request) {
	s.remove(w, r, new(Event), "event", "eventRemoveConfirmation.html", "eventRemove.html")
}

func normaliseDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, location)
}
