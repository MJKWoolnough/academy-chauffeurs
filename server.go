package main

import (
	"html/template"

	"github.com/MJKWoolnough/pagination"
	"github.com/MJKWoolnough/store"
)

type Server struct {
	db         *store.Store
	pages      *template.Template
	pagination pagination.Config
}

func NewServer(db *store.Store) (*Server, error) {
	t, err := template.New("templates").ParseGlob("templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Server{
		db:         db,
		pages:      t,
		pagination: pagination.New(),
	}, nil
}
