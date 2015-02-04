package main

import (
	"net/http"

	"github.com/MJKWoolnough/form"
	"github.com/MJKWoolnough/pagination"
	"github.com/MJKWoolnough/store"
)

const clientsPerPage = 20

type Client struct {
	ID                              int
	Name, Reference, Address, Phone string
	Company                         Company
}

func (c *Client) Get() store.TypeMap {
	return store.TypeMap{
		"id":        &c.ID,
		"companyID": &c.Company.ID,
		"name":      &c.Name,
		"ref":       &c.Reference,
		"address":   &c.Address,
		"phone":     &c.Phone,
	}
}

func (c *Client) ParserList() form.ParserList {
	return form.ParserList{
		"id":          form.Int{&c.ID},
		"companyName": form.String{&c.Company.Name},
		"name":        form.String{&c.Name},
		"reference":   form.String{&c.Reference},
		"address":     form.String{&c.Address},
		"phone":       form.String{&c.Phone},
	}
}

func (c *Client) GetCompanyDetails(db *store.Store) {
	if c.Company.ID == 0 || c.Company.Name != "" {
		return
	}
	db.Get(&c.Company)
}

func (c *Client) GetCompanyID(db *store.Store) {
	if c.Company.Name == "" || c.Company.ID > 0 {
		return
	}
	db.Search([]store.Interface{&c.Company}, 0, store.MatchString("name", c.Company.Name))
}

func (Client) Key() string {
	return "id"
}

func (Client) TableName() string {
	return "clients"
}

type clientErrors struct {
	Client
	NameError, CompanyNameError, AddressError, ReferenceError, PhoneError string
}

type ClientListPageVars struct {
	Clients []Client
	pagination.Pagination
}

func (s *Server) clients(w http.ResponseWriter, r *http.Request) {
	clients := make([]Client, clientsPerPage)
	data := make([]store.Interface, clientsPerPage)
	for i := 0; i < clientsPerPage; i++ {
		data[i] = &clients[i]
	}
	s.list(w, r, data, "clients.html", func(n int, p pagination.Pagination) interface{} {
		for i := 0; i < n; i++ {
			clients[i].GetCompanyDetails(s.db)
		}
		return ClientListPageVars{
			clients[:n],
			p,
		}
	})
}

func (s *Server) addClient(w http.ResponseWriter, r *http.Request) {
	var c clientErrors
	s.add(
		w,
		r,
		&c,
		func() bool {
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
			c.GetCompanyID(s.db)
			if c.Company.ID == 0 {
				good = false
				c.CompanyNameError = "Unknown Company"
			}
			return good
		},
		func() {
			c.GetCompanyDetails(s.db)
		},
		"clients",
		"clientAdd.html",
	)
}

func (s *Server) removeClient(w http.ResponseWriter, r *http.Request) {
	var c Client
	s.remove(w, r, &c, "clients", "clientRemoveConfirmation.html", "clientRemove.html")
}

func (s *Server) updateClient(w http.ResponseWriter, r *http.Request) {
	var c clientErrors
	s.update(
		w,
		r,
		&c,
		func() bool {
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
			c.GetCompanyID(s.db)
			if c.Company.ID == 0 {
				good = false
				c.CompanyNameError = "Unknown Company"
			}
			return good
		},
		func() {
			c.GetCompanyDetails(s.db)
		},
		"clients",
		"clientUpdate.html",
	)
}

func (s *Server) autocompleteClientName(w http.ResponseWriter, r *http.Request) {
	clients := make([]Client, clientsPerPage)
	data := make([]store.Interface, clientsPerPage)
	for i := 0; i < clientsPerPage; i++ {
		data[i] = &clients[i]
	}
	s.autocomplete(w, r, data, "name")
}
