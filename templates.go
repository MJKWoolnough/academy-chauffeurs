package main

import (
	"net/http"

	"github.com/MJKWoolnough/form"
	"github.com/MJKWoolnough/store"
)

type parserStore interface {
	form.Parser
	store.Interface
}

func (s *Server) list(w http.ResponseWriter, r *http.Request, d []store.Interface, t string, v func(int, Pagination) interface{}) {
	var page uint
	r.ParseForm()
	form.Parse(form.Single{"page": form.Uint{&page}}, r.Form)
	num, err := s.db.Count(d[0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	maxPage := num / len(d)
	if num%len(d) > 0 {
		maxPage++
	}
	if page > maxPage {
		page = maxPage
	}
	n, err := s.db.GetPage(d, int(page*p))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.pages.ExecuteTemplate(w, t, v(n, Pagination{page, uint(maxPage)}))
}

func (s *Server) add(w http.ResponseWriter, r *http.Request, f parserStore, v func() bool, redirect, template string) {
	r.ParseForm()
	if r.Method == "POST" {
		form.Parse(f, r.PostForm)
		if v() {
			n, err := s.db.Set(f)
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
	if r.PostForm["confirm"] != "" {
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
			n, err := s.db.Set(f)
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
