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

func (e *Event) GetClientName(db *store.Store) string {
	if e.ClientID == 0 || e.ClientName != "" {
		return e.ClientName
	}
	db.Get(&e.Client)
	return e.ClientName
}

func (e *Event) GetClientID(db *store.Store) int {
	if e.ClientName == "" || e.ClientID > 0 {
		return e.ClientID
	}
	var cli Client
	db.Search([]store.Interface{&cli}, 0, store.MatchString("name", e.ClientName))
	e.ClientID = cli.ID
	return cli.ID
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
	s.eventList(w, r, t, ModeNormal, nil)
}

type EventErrors struct {
	Event
	ClientName                                   string
	FromError, ToError, ClientError, DriverError string
}

func (e *EventErrors) ParserList() form.ParserList {
	pl := e.Event.ParserList()
	pl["clientName"] = form.String{&e.ClientName}
	return pl
}

func (s *Server) addEvent(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	s.addUpdateEvent(w, r, Event{}, 0)
}

func (s *Server) addUpdateEvent(w http.ResponseWriter, r *http.Request, ev Event, forcePage int) {
	var date time.Time
	e := EventErrors{Event: ev}
	form.Parse(&e, r.PostForm)
	form.ParseValue("clientName", form.String{&e.ClientName}, r.PostForm)
	if e.Start.IsZero() {
		forcePage = 1
		e.Start = time.Now()
		date = normaliseDay(time.Now())
	} else {
		date = normaliseDay(e.Start)
	}

	//check dates are valid for driver

	if forcePage == 1 {
		e.Start = normaliseDay(e.Start)
		s.eventList(w, r, date, ModeStart, &e.Event)
		return
	}
	form.ParseValue("date", form.TimeFormat{&date, dateFormat}, r.PostForm)
	if e.End.IsZero() || !e.End.After(e.Start) || forcePage == 2 {
		e.End = normaliseDay(e.Start)
		s.eventList(w, r, date, ModeEnd, &e.Event)
		return
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
		_, err = s.db.Search([]store.Interface{&e.Driver}, 0, store.MatchString("name", e.Driver.Name))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if e.Client.ID == 0 {
			good = false
			e.ClientError = "Unknown Client Name"
		}
		if e.Driver.ID == 0 {
			good = false
			e.DriverError = "Unknown Driver Name"
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
		e.GetClientName(s.db)
		e.GetDriverDetails(s.db)
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
	s.addUpdateEvent(w, r, e, forcePage)
}

func (s *Server) removeEvent(w http.ResponseWriter, r *http.Request) {
	s.remove(w, r, new(Event), "event", "eventRemoveConfirmation.html", "eventRemove.html")
}

func normaliseDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, location)
}
