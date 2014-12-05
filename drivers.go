package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/MJKWoolnough/form"
	"github.com/MJKWoolnough/store"
)

const driversPerPage = 20

type driverErrors struct {
	Driver
	NameError, RegistrationError, PhoneError string
}

type DriverListPageVars struct {
	Drivers []Driver
	Pagination
}

func (s *Server) drivers(w http.ResponseWriter, r *http.Request) {
	var page uint
	r.ParseForm()
	form.Parse(form.Single{"page", form.Uint{&page}}, r.Form)
	num, err := s.db.Count(new(Driver))
	maxPage := num / driversPerPage
	if num%driversPerPage > 0 {
		maxPage++
	}
	drivers := make([]Driver, driversPerPage)
	data := make([]store.Interface, driversPerPage)
	for i := 0; i < driversPerPage; i++ {
		data[i] = &drivers[i]
	}
	_, err = s.db.GetPage(data, int(page)*driversPerPage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	vars := DriverListPageVars{drivers, Pagination{page, make([]struct{}, maxPage)}}
	s.pages.ExecuteTemplate(w, "drivers.html", &vars)
}

func (s *Server) addDriver(w http.ResponseWriter, r *http.Request) {
	var d driverErrors
	if r.Method == "POST" {
		r.ParseForm()
		form.Parse(&d, r.PostForm)
		d.ID = 0
		errors := false
		if d.Name == "" {
			errors = true
			d.NameError = "Name required"
		}
		if d.Registration == "" {
			errors = true
			d.RegistrationError = "Registration required"
		}
		if d.Phone == "" {
			errors = true
			d.PhoneError = "Phone Number required"
		}
		if !errors {
			var err error
			d.ID, err = s.db.Set(&d.Driver)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			fmt.Println(d.ID)
			num, err := s.db.Count(new(Driver))
			maxPage := num / driversPerPage
			if num%driversPerPage > 0 {
				maxPage++
			}
			http.Redirect(w, r, "drivers?page="+strconv.Itoa(maxPage), http.StatusFound)
			return
		}
	}
	s.pages.ExecuteTemplate(w, "driverAdd.html", d)
}

func (s *Server) removeDriver(w http.ResponseWriter, r *http.Request) {
	var d driverErrors
	r.ParseForm()
	form.Parse(&d, r.PostForm)
	err := s.db.Get(&d.Driver)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if d.ID == 0 {
		http.Redirect(w, r, "drivers.html", http.StatusFound)
		return
	}
	if r.PostForm.Get("confirm") != "" {
		s.pages.ExecuteTemplate(w, "driverRemoveConfirmation.html", nil)
	} else if err := s.db.Delete(&d.Driver); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		s.pages.ExecuteTemplate(w, "driverRemove.html", d)
	}
}

func (s *Server) updateDriver(w http.ResponseWriter, r *http.Request) {
	var d driverErrors
	r.ParseForm()
	if r.Method == "POST" {
		form.Parse(&d, r.PostForm)
		if d.Name == "" {
			d.NameError = "Name required"
		}
		if d.Registration == "" {
			d.RegistrationError = "Registration required"
		}
		if d.Phone == "" {
			d.PhoneError = "Phone Number required"
		}
	} else if err := s.db.Get(&d); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.pages.ExecuteTemplate(w, "driverUpdate.html", d)
}
