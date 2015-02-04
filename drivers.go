package main

import (
	"net/http"

	"github.com/MJKWoolnough/form"
	"github.com/MJKWoolnough/pagination"
	"github.com/MJKWoolnough/store"
)

const driversPerPage = 20

type Driver struct {
	ID                        int
	Name, Registration, Phone string
}

func (d *Driver) Get() store.TypeMap {
	return store.TypeMap{
		"id":           &d.ID,
		"name":         &d.Name,
		"registration": &d.Registration,
		"phone":        &d.Phone,
	}
}

func (d *Driver) ParserList() form.ParserList {
	return form.ParserList{
		"id":           form.Int{&d.ID},
		"name":         form.String{&d.Name},
		"registration": form.String{&d.Registration},
		"phone":        form.String{&d.Phone},
	}
}

func (Driver) Key() string {
	return "id"
}

func (Driver) TableName() string {
	return "drivers"
}

type driverErrors struct {
	Driver
	NameError, RegistrationError, PhoneError string
}

type DriverListPageVars struct {
	Drivers []Driver
	pagination.Pagination
}

func (s *Server) drivers(w http.ResponseWriter, r *http.Request) {
	drivers := make([]Driver, driversPerPage)
	data := make([]store.Interface, driversPerPage)
	for i := 0; i < driversPerPage; i++ {
		data[i] = &drivers[i]
	}
	s.list(w, r, data, "drivers.html", func(n int, p pagination.Pagination) interface{} {
		return DriverListPageVars{
			drivers[:n],
			p,
		}
	})
}

func (s *Server) addDriver(w http.ResponseWriter, r *http.Request) {
	var d driverErrors
	s.add(w, r, &d, func() bool {
		good := true
		if d.Name == "" {
			good = false
			d.NameError = "Driver Name required"
		}
		if d.Registration == "" {
			good = false
			d.RegistrationError = "Registration required"
		}
		if d.Phone == "" {
			good = false
			d.PhoneError = "Phone required"
		}
		return good
	}, nil, "drivers", "driverAdd.html")
}

func (s *Server) removeDriver(w http.ResponseWriter, r *http.Request) {
	s.remove(w, r, new(Driver), "driver", "driverRemoveConfirmation.html", "driverRemove.html")
}

func (s *Server) updateDriver(w http.ResponseWriter, r *http.Request) {
	var d driverErrors
	s.update(w, r, &d, func() bool {
		good := true
		if d.Name == "" {
			good = false
			d.NameError = "Driver Name required"
		}
		if d.Registration == "" {
			good = false
			d.RegistrationError = "Registration required"
		}
		if d.Phone == "" {
			good = false
			d.PhoneError = "Phone required"
		}
		return good
	}, nil, "drivers", "driverUpdate.html")
}

func (s *Server) autocompleteDriveName(w http.ResponseWriter, r *http.Request) {
	drivers := make([]Driver, driversPerPage)
	data := make([]store.Interface, driversPerPage)
	for i := 0; i < driversPerPage; i++ {
		data[i] = &drivers[i]
	}
	s.autocomplete(w, r, data, "name")
}
