package main

import (
	"net/http"

	"github.com/MJKWoolnough/store"
)

const driversPerPage = 20

type driverErrors struct {
	Driver
	NameError, RegistrationError, PhoneError string
}

type DriverListPageVars struct {
	Drivers            []Driver
	CurrPage, LastPage uint
}

func (s *Server) drivers(w http.ResponseWriter, r *http.Request) {
	drivers := make([]Driver, driversPerPage)
	data := make([]store.Interface, driversPerPage)
	for i := 0; i < driversPerPage; i++ {
		data[i] = &drivers[i]
	}
	s.list(w, r, data, "drivers.html", func(n int, currPage, lastPage uint) interface{} {
		return DriverListPageVars{
			drivers[:n],
			currPage, lastPage,
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
	}, "drivers", "driverAdd.html")
}

func (s *Server) removeDriver(w http.ResponseWriter, r *http.Request) {
	s.remove(w, r, new(Driver), "driver.html", "driverRemoveConfirmation.html", "driverRemove.html")
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
	}, "drivers", "driverUpdate.html")
}
