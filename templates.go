package main

import (
	"net/http"

	"github.com/MJKWoolnough/form"
	"github.com/MJKWoolnough/pagination"
	"github.com/MJKWoolnough/store"
)

type parserStore interface {
	form.ParserLister
	store.Interface
}

type ListVars struct {
	Drivers []store.Interface
	pagination.Pagination
}

func (s *Server) list(w http.ResponseWriter, r *http.Request, d []store.Interface, t string, v func(int, pagination.Pagination) interface{}) {
	var page uint
	r.ParseForm()
	form.Parse(form.Single{"page", form.Uint{&page}}, r.Form)
	if page > 0 {
		page--
	}
	num, err := s.db.Count(d[0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	maxPage := num / len(d)
	if num%len(d) == 0 {
		maxPage--
	}
	if page > uint(maxPage) {
		page = uint(maxPage)
	}
	n, err := s.db.GetPage(d, int(page)*len(d))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = n
	s.pages.ExecuteTemplate(w, t, v(n, s.pagination.Get(page, uint(maxPage))))
}

func (s *Server) add(w http.ResponseWriter, r *http.Request, f parserStore, v func() bool, redirect, template string) {
	r.ParseForm()
	if r.Method == "POST" {
		form.Parse(f, r.PostForm)
		if v() {
			_, err := s.db.Set(f)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, redirect, http.StatusFound)
			return
		}
	}
	s.pages.ExecuteTemplate(w, template, f)
}
func (s *Server) remove(w http.ResponseWriter, r *http.Request, f parserStore, redirect, confirmation, template string) {
	r.ParseForm()
	form.Parse(f, r.PostForm)
	err := s.db.Get(f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if i, ok := f.Get()[f.Key()].(*int); !ok || *i == 0 {
		http.Redirect(w, r, redirect, http.StatusFound)
		return
	}
	if len(r.PostForm["confirm"]) != 0 {
		s.pages.ExecuteTemplate(w, confirmation, nil)
	} else if err = s.db.Delete(f); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		s.pages.ExecuteTemplate(w, template, f)
	}
}

func (s *Server) update(w http.ResponseWriter, r *http.Request, f parserStore, v func() bool, redirect, template string) {
	r.ParseForm()
	if r.Method == "POST" {
		form.Parse(f, r.PostForm)
		if v() {
			_, err := s.db.Set(f)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, redirect, http.StatusFound)
			return
		}
	} else {
		form.Parse(f, r.Form)
		err := s.db.Get(f)
		if err != nil {

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if i, ok := f.Get()[f.Key()].(*int); !ok || *i == 0 {
		http.Redirect(w, r, redirect, http.StatusFound)
		return
	}
	s.pages.ExecuteTemplate(w, template, f)
}
