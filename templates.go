package main

import (
	"encoding/json"
	"net/http"
	"time"

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
	form.Parse(form.ParserList{"page": form.Uint{&page}}, r.Form)
	if page > 0 {
		page--
	}
	num, err := s.db.Count(d[0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	maxPage := num / len(d)
	if num%len(d) == 0 && maxPage > 0 {
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
	s.pages.ExecuteTemplate(w, t, v(n, s.pagination.Get(page, uint(maxPage))))
}

func (s *Server) add(w http.ResponseWriter, r *http.Request, f parserStore, v func() bool, d func(), redirect, template string) {
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
	if d != nil {
		d()
	}
	s.pages.ExecuteTemplate(w, template, f)
}
func (s *Server) remove(w http.ResponseWriter, r *http.Request, f parserStore, redirect, confirmation, template string) {
	if r.Method == "POST" {
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
		if err = s.db.Delete(f); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else if len(r.PostForm["confirm"]) != 0 {
			s.pages.ExecuteTemplate(w, confirmation, nil)
			return
		}
	}
	s.pages.ExecuteTemplate(w, template, f)
}

func (s *Server) update(w http.ResponseWriter, r *http.Request, f parserStore, v func() bool, d func(), redirect, template string) {
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
	if d != nil {
		d()
	}
	s.pages.ExecuteTemplate(w, template, f)
}

func (s *Server) autocomplete(w http.ResponseWriter, r *http.Request, d []store.Interface, column string) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
	r.ParseForm()
	partial := r.PostForm.Get("partial")
	store.Sort(d, column, true)
	n, err := s.db.Search(d, 0, store.Like(column, partial+"%"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if n < len(d) {
		m, err := s.db.Search(d[n:], 0, store.Like(column, "%"+partial+"%"), store.NotLike(column, partial+"%"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		n += m
	}
	type JSONAutocomplete struct {
		Time int64
		Data []string
	}
	var ja JSONAutocomplete
	ja.Data = make([]string, 0, n)
	for i := 0; i < n; i++ {
		if s, ok := d[i].Get()[column].(*string); ok {
			ja.Data = append(ja.Data, *s)
		}
	}
	wrap := r.PostForm.Get("wrap")
	if wrap != "" {
		w.Write([]byte(wrap + "("))
	}
	ja.Time = time.Now().UnixNano()
	json.NewEncoder(w).Encode(ja)
	if wrap != "" {
		w.Write([]byte{')', ';'})
	}
}
