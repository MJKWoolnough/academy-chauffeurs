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
	Driver
	Client
}

func (e *Event) Get() store.TypeMap {
	return store.TypeMap{
		"id":           &e.ID,
		"start":        &e.Start,
		"end":          &e.End,
		"from":         &e.From,
		"to":           &e.To,
		"pickupTime":   &e.Pickup,
		"waitingTime":  &e.Waiting,
		"parkingCosts": &e.Parking,
		"milesDriver":  &e.Miles,
		"clientId":     &e.Client.ID,
		"driverId":     &e.Driver.ID,
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
		"clientName":         form.String{&e.Client.Name},
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

func (e *Event) GetClientDetails(db *store.Store) {
	if e.Client.ID == 0 || e.Client.Name != "" {
		return
	}
	db.Get(&e.Client)
}

func (e *Event) GetClientID(db *store.Store) {
	if e.Client.Name == "" || e.Client.ID > 0 {
		return
	}
	db.Search([]store.Interface{&e.Client}, 0, store.MatchString("name", e.Client.Name))
}

func (e *Event) GetDriverDetails(db *store.Store) {
	if e.Driver.ID == 0 || e.Driver.Name != "" {
		return
	}
	db.Get(&e.Driver)
}

func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	var t time.Time
	r.ParseForm()
	form.ParseValue("date", form.TimeFormat{&t, dateFormat}, r.Form)
	if t.IsZero() {
		t = time.Now()
	}
	s.eventList(w, r, t, ModeNormal, nil, false)
}

type EventErrors struct {
	Event
	FromError, ToError, ClientError string
}

func (s *Server) addEvent(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	s.addUpdateEvent(w, r, Event{}, 0, false)
}

func (s *Server) addUpdateEvent(w http.ResponseWriter, r *http.Request, ev Event, forcePage int, isUpdate bool) {
	var date time.Time
	e := EventErrors{Event: ev}
	form.Parse(&e, r.PostForm)
	if e.Start.IsZero() {
		forcePage = 1
		e.Start = time.Now()
		date = normaliseDay(time.Now())
	} else {
		date = normaliseDay(e.Start)
	}

	form.ParseValue("date", form.TimeFormat{&date, dateFormat}, r.PostForm)

	if forcePage == 1 {
		e.Start = normaliseDay(e.Start)
		s.eventList(w, r, date, ModeStart, &e.Event, isUpdate)
		return
	}

	e.GetDriverDetails(s.db)
	if e.Driver.ID == 0 { //bad driver
		http.Redirect(w, r, "/events", http.StatusFound)
	}

	if n, _ := s.db.SearchCount(new(Event),
		store.NotMatchInt("id", e.ID),
		store.GreaterThanEqual("start", int(e.Event.Start.Unix())),
		store.LessThanEqual("end", int(e.Event.Start.Unix()))); n > 0 { //start in invalid position
		http.Redirect(w, r, "/events", http.StatusFound)
	}

	if e.End.IsZero() || !e.End.After(e.Start) || forcePage == 2 {
		e.End = normaliseDay(e.Start)
		s.eventList(w, r, date, ModeEnd, &e.Event, isUpdate)
		return
	}

	if n, _ := s.db.SearchCount(new(Event),
		store.GreaterThanEqual("start", int(e.Event.End.Unix())),
		store.LessThanEqual("end", int(e.Event.End.Unix()))); n > 0 { //start in invalid position
		http.Redirect(w, r, "/events", http.StatusFound)
	}
	if r.PostForm.Get("submit") != "" {
		good := true
		if e.From == "" {
			good = false
			e.FromError = "From Address Required"
		}
		if e.To == "" {
			good = false
			e.ToError = "Destination Address Required"
		}
		_, err := s.db.Search([]store.Interface{&e.Client}, 0, store.MatchString("name", e.Client.Name))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		e.GetClientID(s.db)
		if e.Client.ID == 0 {
			good = false
			e.ClientError = "Unknown Client Name"
		}
		if good {
			_, err := s.db.Set(&e)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "events?date="+e.Start.Format(dateFormat), http.StatusFound)
			return
		}
	} else {
		e.GetClientDetails(s.db)
	}
	s.pages.ExecuteTemplate(w, "eventEditDetails.html", e)
}

func (s *Server) updateEvent(w http.ResponseWriter, r *http.Request) {
	var (
		e         Event
		forcePage int
	)
	r.ParseForm()
	form.ParseValue("id", form.Int{&e.ID}, r.PostForm)
	if e.ID == 0 {
		http.Redirect(w, r, "events", http.StatusFound)
		return
	}
	err := s.db.Get(&e)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if e.ID == 0 {
		http.Redirect(w, r, "events", http.StatusFound)
		return
	}
	form.ParseValue("force", form.Int{&forcePage}, r.PostForm)
	s.addUpdateEvent(w, r, e, forcePage, true)
}

func (s *Server) removeEvent(w http.ResponseWriter, r *http.Request) {
	s.remove(w, r, new(Event), "event", "eventRemoveConfirmation.html", "eventRemove.html")
}

func normaliseDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, location)
}
