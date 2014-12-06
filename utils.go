package main

import (
	"html/template"
	"strconv"
)

const (
	endShow  = 3
	surround = 3
)

type Pagination struct {
	CurrPage, NumPages uint
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
		} else if page < endShow || page >= p.NumPages-endShow || (int(page) >= int(p.CurrPage)-surround && page <= p.CurrPage+surround) || (p.CurrPage-surround-1 == endShow && page == endShow) || (p.CurrPage+surround+1 == p.NumPages-endShow-1 && page == p.NumPages-endShow-1) {
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
