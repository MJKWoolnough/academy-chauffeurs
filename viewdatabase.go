package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/MJKWoolnough/form"
	"github.com/MJKWoolnough/pagination"
	"github.com/MJKWoolnough/store"
)

type viewVars struct {
	Data []store.Interface
	pagination.Pagination
}

var viewTemplate *template.Template

const perPage = 10

func init() {
	viewTemplate = template.Must(template.New("viewDatabase").Parse(`<html>
	<body>
		<table>
			<tr>
{{range (index .Data 0).Get}}				<th>{{.}}</th>
{{end}}
			</tr>
{{_, $data := range .Data}}				<tr>
{{range .Get}}				<td>{{.}}</td>
			</tr>
{{end}}
		</table>
		{{.Pagination.HTML "?page="}}
	</body>
</html>
`))
}

func setupView(s *Server) {
	http.HandleFunc("/databaseDrivers", s.viewDrivers)
}

func (s *Server) databaseDisplay(w http.ResponseWriter, r *http.Request, data []store.Interface) {
	var page uint
	r.ParseForm()
	form.ParseValue("page", form.Uint{&page}, r.Form)
	if page > 0 {
		page--
	}
	num, err := s.db.Count(data[0])
	maxPage := uint(num / len(data))
	if num%len(data) == 0 && maxPage > 0 {
		maxPage--
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	num, err = s.db.GetPage(data, int(page)*len(data))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	data = data[:num]
	err = viewTemplate.Execute(w, viewVars{data, s.pagination.Get(page, maxPage)})
	if err != nil {
		fmt.Println(err)
	}
}

func (s *Server) viewDrivers(w http.ResponseWriter, r *http.Request) {
	var data [perPage]Driver
	iData := make([]store.Interface, perPage)
	for i := 0; i < perPage; i++ {
		iData[i] = &data[i]
	}
	s.databaseDisplay(w, r, iData)
}
