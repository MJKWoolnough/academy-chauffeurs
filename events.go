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
		"start":              form.Time{&e.Start},
		"end":                form.Time{&e.End},
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

type EventTemplateVars struct {
	PrevStr, NowStr, NextStr string
	Drivers                  []Driver
	DriverEvents             [][]Event
}

func (e *EventTemplateVars) BlockFilled(driver, time int) bool {
	for _, event := range e.DriverEvents[driver] {
		if time >= int(event.Start.Unix()) && time <= int(event.End.Unix()) {
			return true
		}
	}
	return false
}

func (e *EventTemplateVars) BlockInfo(driver, time int) *Event {
	for _, e := range e.DriverEvents[driver] {
		if time >= int(e.Start.Unix()) && time <= int(e.End.Unix()) {
			return &e
		}
	}
	return nil
}

const dateFormat = "2006-01-02"

var location *time.Location

func init() {
	location, _ = time.LoadLocation("") // "" == UTC
}

func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	var (
		t time.Time
		e EventTemplateVars
	)
	form.Parse(form.ParserList{"date": form.TimeFormat{&t, dateFormat}}, r.Form)
	if t.Unix() == 0 {
		t = time.Now()
	}
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, location)
	e.PrevStr = t.AddDate(0, 0, -1).Format(dateFormat)
	e.NowStr = t.Format(dateFormat)
	e.NextStr = t.AddDate(0, 0, 1).Format(dateFormat)
	numDrivers, err := s.db.SearchCount(new(Driver))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	e.Drivers = make([]Driver, numDrivers)
	driversI := make([]store.Interface, numDrivers)
	for i := 0; i < numDrivers; i++ {
		driversI[i] = &e.Drivers[i]
	}
	_, err = s.db.Search(store.Sort(driversI, "name", true), 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	e.DriverEvents = make([][]Event, numDrivers)
	startTime := int(t.Unix())
	endTime := int(t.AddDate(0, 0, 1).Unix())
	for i := 0; i < numDrivers; i++ {
		searchers := []store.Searcher{
			store.MatchInt("driverId", e.Drivers[i].ID),
			store.Or(
				store.Between("start", startTime, endTime),
				store.Between("end", startTime, endTime),
			),
		}
		numEvents, err := s.db.SearchCount(new(Event), searchers...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		e.DriverEvents[i] = make([]Event, numEvents)
		eventsI := make([]store.Interface, numEvents)
		for j := 0; j < numEvents; j++ {
			eventsI[j] = &e.DriverEvents[i][j]
		}
		_, err = s.db.Search(eventsI, 0, searchers...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	s.pages.ExecuteTemplate(w, "events.html", &e)
}

func (s *Server) addEvent(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) updateEvent(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) removeEvent(w http.ResponseWriter, r *http.Request) {
	s.remove(w, r, new(Event), "event", "eventRemoveConfirmation.html", "eventRemove.html")
}
