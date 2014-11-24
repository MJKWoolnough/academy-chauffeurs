package main

import (
	"html/template"

	"github.com/MJKWoolnough/store"
)

type Server struct {
	db    *store.Store
	pages map[string]*template.Template
}

func NewServer(db *store.Store) *Server {
	return &Server{
		db: db,
	}
}
