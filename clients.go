package main

import (
	"net/http"

	"github.com/MJKWoolnough/store"
)

const clientsPerPage = 20

type clientErrors struct {
	Client
	NameError, CompanyNameError, AddressError, ReferenceError, PhoneError string
}

type ClientListPageVars struct {
	Clients            []Client
	CurrPage, LastPage uint
}

func (s *Server) clients(w http.ResponseWriter, r *http.Request) {
	clients := make([]Client, clientsPerPage)
	data := make([]store.Interface, clientsPerPage)
	for i := 0; i < clientsPerPage; i++ {
		data[i] = &clients[i]
	}
	s.list(w, r, data, "clients.html", func(n int, currPage, lastPage uint) interface{} {
		for i := 0; i < n; i++ {
			clients[i].companyName(s.db)
		}
		return ClientListPageVars{
			clients[:n],
			currPage, lastPage,
		}
	})
}

func (s *Server) addClient(w http.ResponseWriter, r *http.Request) {
	var c clientErrors
	s.add(w, r, &c, func() bool {
		good := true
		if c.Name == "" {
			good = false
			c.NameError = "Client Name required"
		}
		if c.Address == "" {
			good = false
			c.AddressError = "Address required"
		}
		if c.Reference == "" {
			good = false
			c.ReferenceError = "Reference required"
		}
		if c.Phone == "" {
			good = false
			c.PhoneError = "Valid Phone Number Required"
		}
		c.companyID(s.db)
		if c.CompanyID == 0 {
			good = false
			c.CompanyNameError = "Unknown Company"
		}
		return good
	}, "clients", "clientAdd.html")
}

func (s *Server) removeClient(w http.ResponseWriter, r *http.Request) {
	var c Client
	s.remove(w, r, &c, "clients.html", "clientRemoveConfirmation.html", "clientRemove.html")
}

func (s *Server) updateClient(w http.ResponseWriter, r *http.Request) {
	var c clientErrors
	s.update(w, r, &c, func() bool {
		good := true
		if c.Name == "" {
			good = false
			c.NameError = "Client Name required"
		}
		if c.Address == "" {
			good = false
			c.AddressError = "Address required"
		}
		if c.Reference == "" {
			good = false
			c.ReferenceError = "Reference required"
		}
		if c.Phone == "" {
			good = false
			c.PhoneError = "Phone Number Required"
		}
		c.companyID(s.db)
		if c.CompanyID == 0 {
			good = false
			c.CompanyNameError = "Unknown Company"
		}
		return good
	}, "clients", "clientUpdate.html")
}
