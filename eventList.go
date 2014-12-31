package main

import (
	"net/http"
	"time"

	"github.com/MJKWoolnough/form"
	"github.com/MJKWoolnough/store"
)

func (e *Event) NumBlocks() time.Duration {
	return e.End.Sub(e.Start) / blockDuration
}

type EventTemplateVars struct {
	today        time.Time
	Drivers      []Driver
	DriverEvents [][]Event
}

func (e *EventTemplateVars) BlockFilled(driver int, t int) bool {
	for _, event := range e.DriverEvents[driver] {
		if t >= int(event.Start.Unix()) && t <= int(event.End.Unix()) {
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

func (e *EventTemplateVars) Date(d int) string {
	e.today.AddDate(0, 0, d).Format(dateFormat)
}

func (e *EventTemplateVars) BlockTimes() []time.Time {
	t := e.today.AddDate(0, 0, 0)
	tomorrow := e.today.AddDate(0, 0, 1)
	times := make([]time.Time, 0, time.Hour*24/blockDuration)
	for t.Before(tomorrow) {
		times = append(times, t)
		t = t.Add(blockDuration)
	}
	return times
}

const (
	blockDuration = time.Minute * 15
	dateFormat    = "2006-01-02"
	timeFormat    = "15:04"
)

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
	e.today = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, location)
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
