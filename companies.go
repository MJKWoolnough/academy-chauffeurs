package main

import (
	"net/http"

	"github.com/MJKWoolnough/pagination"
	"github.com/MJKWoolnough/store"
)

const companiesPerPage = 20

type companyErrors struct {
	Company
	NameError, AddressError, ReferenceError string
}

type CompanyListPageVars struct {
	Companies []Company
	pagination.Pagination
}

func (s *Server) companies(w http.ResponseWriter, r *http.Request) {
	companies := make([]Company, companiesPerPage)
	data := make([]store.Interface, companiesPerPage)
	for i := 0; i < companiesPerPage; i++ {
		data[i] = &companies[i]
	}
	s.list(w, r, data, "companies.html", func(n int, p pagination.Pagination) interface{} {
		return CompanyListPageVars{
			companies[:n],
			p,
		}
	})
}

func (s *Server) addCompany(w http.ResponseWriter, r *http.Request) {
	var c companyErrors
	s.add(w, r, &c, func() bool {
		good := true
		if c.Name == "" {
			good = false
			c.NameError = "Company Name required"
		}
		if c.Address == "" {
			good = false
			c.AddressError = "Address required"
		}
		if c.Reference == "" {
			good = false
			c.ReferenceError = "Reference required"
		}
		return good
	}, "companies", "companyAdd.html")
}

func (s *Server) removeCompany(w http.ResponseWriter, r *http.Request) {
	var c Company
	s.remove(w, r, &c, "companies", "companyRemoveConfirmation.html", "companyRemove.html")
}

func (s *Server) updateCompany(w http.ResponseWriter, r *http.Request) {
	var c companyErrors
	s.update(w, r, &c, func() bool {
		good := true
		if c.Name == "" {
			good = false
			c.NameError = "Company Name required"
		}
		if c.Address == "" {
			good = false
			c.AddressError = "Address required"
		}
		if c.Reference == "" {
			good = false
			c.ReferenceError = "Reference required"
		}
		return good
	}, "companies", "companyUpdate.html")
}
