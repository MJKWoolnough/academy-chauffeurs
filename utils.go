package main

import (
	"html/template"
	"strconv"
)

const (
	defaultEndShow  = 3
	defaultSurround = 3
)

type Pagination struct {
	CurrPage, NumPages uint
	Surround, Ends     uint
}

func NewPagination(current, total uint) Pagination {
	return Pagination{
		CurrPage: current,
		NumPages: total,
		Surround: defaultSurround,
		Ends:     defaultEndShow,
	}
}

func (p Pagination) HTML() template.HTML {
	if p.NumPages <= 1 {
		return ""
	}
	html := "<div class=\"pagination\">"
	if p.CurrPage == 0 {
		html += "<a>Prev</a> "
	} else {
		html += "<a href=\"?page=" + strconv.Itoa(int(p.CurrPage)-1) + "\">Prev</a> "
	}
	elipses := true
	for page := uint(0); page < p.NumPages; page++ {
		numStr := strconv.Itoa(int(page + 1))
		if page == p.CurrPage {
			html += "<a>" + numStr + "</a> "
		} else if page < p.Ends || page >= p.NumPages-p.Ends || (int(page) >= int(p.CurrPage)-int(p.Surround) && page <= p.CurrPage+p.Surround) || (p.CurrPage-p.Surround-1 == p.Ends && page == p.Ends) || (p.CurrPage+p.Surround+1 == p.NumPages-p.Ends-1 && page == p.NumPages-p.Ends-1) {
			html += "<a href=\"?page=" + numStr + "\">" + numStr + "</a> "
			elipses = true
		} else if elipses {
			elipses = false
			html += "... "
		}
	}

	if p.CurrPage == p.NumPages-1 {
		html += "<a>Next</a> "
	} else {
		html += "<a href=\"?page=" + strconv.Itoa(int(p.CurrPage)+1) + "\">Next</a>"
	}
	html += "</div>"
	return template.HTML(html)
}
