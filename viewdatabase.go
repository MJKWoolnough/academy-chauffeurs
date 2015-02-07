package main

import (
	"fmt"
	"net/http"

	"github.com/MJKWoolnough/form"
	"github.com/MJKWoolnough/pagination"
	"github.com/MJKWoolnough/store"
)

type viewVars struct {
	Data []store.Interface
	pagination.Pagination
}

const perPage = 10

func setupView(s *Server) {
	http.HandleFunc("/databaseDrivers", s.viewDrivers)
	http.HandleFunc("/databaseCompanies", s.viewCompanies)
	http.HandleFunc("/databaseClients", s.viewClients)
	http.HandleFunc("/databaseEvents", s.viewEvents)
}

func (s *Server) databaseDisplay(w http.ResponseWriter, r *http.Request, data []store.Interface) {
	var page uint
	r.ParseForm()
	form.ParseValue("page", form.Uint{&page}, r.Form)
	if page > 0 {
		page--
	}
	num, err := s.db.Count(data[0])
	maxPage := uint(num / len(data))
	if num%len(data) == 0 && maxPage > 0 {
		maxPage--
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	num, err = s.db.GetPage(data, int(page)*len(data))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	data = data[:num]
	err = s.pages.ExecuteTemplate(w, "viewDatabase.html", viewVars{data, s.pagination.Get(page, maxPage)})
	if err != nil {
		fmt.Println(err)
	}
}

func (s *Server) viewDrivers(w http.ResponseWriter, r *http.Request) {
	var data [perPage]Driver
	iData := make([]store.Interface, perPage)
	for i := 0; i < perPage; i++ {
		iData[i] = &data[i]
	}
	s.databaseDisplay(w, r, iData)
}

func (s *Server) viewCompanies(w http.ResponseWriter, r *http.Request) {
	var data [perPage]Company
	iData := make([]store.Interface, perPage)
	for i := 0; i < perPage; i++ {
		iData[i] = &data[i]
	}
	s.databaseDisplay(w, r, iData)
}

func (s *Server) viewClients(w http.ResponseWriter, r *http.Request) {
	var data [perPage]Client
	iData := make([]store.Interface, perPage)
	for i := 0; i < perPage; i++ {
		iData[i] = &data[i]
	}
	s.databaseDisplay(w, r, iData)
}

func (s *Server) viewEvents(w http.ResponseWriter, r *http.Request) {
	var data [perPage]Event
	iData := make([]store.Interface, perPage)
	for i := 0; i < perPage; i++ {
		iData[i] = &data[i]
	}
	s.databaseDisplay(w, r, iData)
}
