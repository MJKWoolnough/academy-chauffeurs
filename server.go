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
	t := template.New("templates")
	t.Funcs(template.FuncMap{"pagination": paginationHTML})
	_, err := t.ParseGlob("templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Server{
		db:    db,
		pages: t,
	}, nil
}
