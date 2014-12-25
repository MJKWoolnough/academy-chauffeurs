package main

import (
	"net/http"
	"time"

	"github.com/MJKWoolnough/form"
	"github.com/MJKWoolnough/store"
)

type Event struct {
	ID        int
	Start     time.Time
	From, To  string
	EDuration int64
	Pickup    time.Time
	Waiting   int64
	Parking   int
	Miles     int
	ClientId  int
	Driver
}

func (e *Event) Get() store.TypeMap {
	return store.TypeMap{
		"id":                 &e.ID,
		"start":              &e.Start,
		"from":               &e.From,
		"to":                 &e.To,
		"estimatedDuration":  &e.EDuration,
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
		"from":               form.String{&e.From},
		"to":                 form.String{&e.To},
		"estimatedDuration":  form.Int64{&e.EDuration},
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

}

func (s *Server) removeEvent(w http.ResponseWriter, r *http.Request) {
	s.remove(w, r, new(Event), "event", "eventRemoveConfirmation.html", "eventRemove.html")
}
