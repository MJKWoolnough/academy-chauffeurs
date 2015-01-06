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

func (s *Server) addEvent(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) updateEvent(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) removeEvent(w http.ResponseWriter, r *http.Request) {
	s.remove(w, r, new(Event), "event", "eventRemoveConfirmation.html", "eventRemove.html")
}
