package main

import (
	"html/template"

	"github.com/MJKWoolnough/pagination"
	"github.com/MJKWoolnough/store"
)

var p = pagination.New()

func paginationHTML(url string, currPage, lastPage uint) template.HTML {
	return template.HTML(p.Get(currPage, lastPage).HTML(url))
}

type Server struct {
	db    *store.Store
	pages *template.Template
}

func NewServer(db *store.Store) (*Server, error) {
	t, err := template.New("templates").ParseGlob("templates/*.html")
	if err != nil {
		return nil, err
	}
	t.Funcs(template.FuncMap{"pagination": paginationHTML})
	return &Server{
		db:    db,
		pages: t,
	}, nil
}
