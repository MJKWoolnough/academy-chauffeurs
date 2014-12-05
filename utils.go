package main

import (
	"html/template"
	"strconv"
)

type Pagination struct {
	CurrPage, NumPages uint
}

func (p Pagination) HTML() template.HTML {
	if p.NumPages <= 1 {
		return ""
	}
	html := "<div class=\"pagination\">"
	if p.CurrPage == 1 {
		html += "<a>Prev</a> "
	} else {
		html += "<a href=\"?page=" + strconv.Itoa(int(p.CurrPage)-1) + "\">Prev</a>"
	}
	elipses := true
	for page := 1; page <= p.NumPages; page++ {
		if page < 3 || page > p.NumPages-3 || page >= p.CurrPage-3 || page <= p.CurrPage+3 || p.CurrPage-3 <= 4 || p.CurrPage+3 >= p.NumPages-4 {
			numStr := strconv.Itoa(page)
			html += "<a href=\"?page=" + numStr + "\">" + numStr + "</a> "
			elipses = true
		} else if elipses {
			elipses = false
			out += "..."
		}
	}

	if p.CurrPage == p.NumPages {
		html += "<a>Next</a> "
	} else {
		html += "<a href=\"?page=" + strconv.Itoa(int(p.CurrPage)+1) + "\">Next</a>"
	}
	html += "</div>"
}
