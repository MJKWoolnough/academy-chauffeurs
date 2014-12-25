package main

import (
	"net/http"

	"github.com/MJKWoolnough/form"
	"github.com/MJKWoolnough/pagination"
	"github.com/MJKWoolnough/store"
)

const clientsPerPage = 20

type Client struct {
	ID, CompanyID                                int
	Name, CompanyName, Reference, Address, Phone string
	db                                           *store.Store
}

func (c *Client) Get() store.TypeMap {
	return store.TypeMap{
		"id":        &c.ID,
		"companyID": &c.CompanyID,
		"name":      &c.Name,
		"ref":       &c.Reference,
		"address":   &c.Address,
		"phone":     &c.Phone,
	}
}

func (c *Client) ParserList() form.ParserList {
	return form.ParserList{
		"id":          form.Int{&c.ID},
		"companyName": form.String{&c.CompanyName},
		"name":        form.String{&c.Name},
		"reference":   form.String{&c.Reference},
		"address":     form.String{&c.Address},
		"phone":       form.String{&c.Phone},
	}
}

func (c *Client) GetCompanyName() string {
	if c.CompanyID == 0 || c.CompanyName != "" {
		return c.CompanyName
	}
	comp := Company{ID: c.CompanyID}
	c.db.Get(&comp)
	c.CompanyName = comp.Name
	return c.CompanyName
}

func (c *Client) GetCompanyID() int {
	if c.CompanyName == "" || c.CompanyID > 0 {
		return c.CompanyID
	}
	var comp Company
	c.db.Search([]store.Interface{&comp}, 0, store.MatchString("name", c.CompanyName))
	c.CompanyID = comp.ID
	return c.CompanyID
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
			clients[i].db = s.db
			clients[i].GetCompanyName()
		}
		return ClientListPageVars{
			clients[:n],
			p,
		}
	})
}

func (s *Server) addClient(w http.ResponseWriter, r *http.Request) {
	c := clientErrors{Client: Client{db: s.db}}
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
		c.GetCompanyID()
		if c.CompanyID == 0 {
			good = false
			c.CompanyNameError = "Unknown Company"
		}
		return good
	}, "clients", "clientAdd.html")
}

func (s *Server) removeClient(w http.ResponseWriter, r *http.Request) {
	var c Client
	s.remove(w, r, &c, "clients", "clientRemoveConfirmation.html", "clientRemove.html")
}

func (s *Server) updateClient(w http.ResponseWriter, r *http.Request) {
	c := clientErrors{Client: Client{db: s.db}}
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
		c.GetCompanyID()
		if c.CompanyID == 0 {
			good = false
			c.CompanyNameError = "Unknown Company"
		}
		return good
	}, "clients", "clientUpdate.html")
}

func (s *Server) autocompleteClientName(w http.ResponseWriter, r *http.Request) {
	clients := make([]Client, clientsPerPage)
	data := make([]store.Interface, clientsPerPage)
	for i := 0; i < clientsPerPage; i++ {
		data[i] = &clients[i]
	}
	s.autocomplete(w, r, data, "name")
}
