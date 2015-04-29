package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/MJKWoolnough/form"
	"github.com/MJKWoolnough/memio"
	"github.com/tealeg/xlsx"
)

func (c *Calls) export(w http.ResponseWriter, r *http.Request) {
	switch r.PostFormValue("type") {
	case "driverEvents":
		c.exportDriverEvents(w, r)
	case "clientEvents":
		c.exportClientEvents(w, r)
	case "companyEvents":
		c.exportCompanyEvents(w, r)
	case "companyClients":
		c.exportCompanyClients(w, r)
	case "clientList":
		c.exportClientList(w, r)
	case "companyList":
		c.exportCompanyList(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (e *EventsFilter) ParserList() form.ParserList {
	return form.ParserList{
		"id":        form.Int64{&e.ID},
		"startTime": form.Int64{&e.Start},
		"endTime":   form.Int64{&e.End},
	}
}

func formatDateTime(msec int64) string {
	return time.Unix(msec/1000, (msec%1000)*1000000).Format("02/01/2006 15:04")
}

func formatDate(msec int64) string {
	return time.Unix(msec/1000, (msec%1000)*1000000).Format("02/01/2006")
}

func formatTime(msec int64) string {
	return time.Unix(msec/1000, (msec%1000)*1000000).Format("15:04")
}

func formatMoney(pence int64) string {
	return strconv.FormatFloat(float64(pence)/100, 'f', 2, 64)
}

func (c *Calls) exportDriverEvents(w http.ResponseWriter, r *http.Request) {
	var f EventsFilter
	err := form.Parse(&f, r.PostForm)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if e, ok := err.(form.Errors); ok {
			for k, v := range e {
				fmt.Fprintf(w, "%s = %s\n", k, v)
			}
		} else {
			w.Write([]byte(err.Error()))
		}
		return
	}
	if f.Start > f.End {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid times"))
		return
	}
	dateStr := formatDate(f.Start)
	if f.Start != f.End {
		dateStr += " to " + formatDate(f.End)
	}
	f.End += 24 * 3600 * 1000
	var (
		e []Event
		d Driver
	)
	err = c.DriverEvents(f, &e)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	err = c.GetDriver(f.ID, &d)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	xls := xlsx.NewFile()
	sheet := xls.AddSheet("Driver Events")
	sheet.AddRow().WriteSlice(&[]string{"Driver Sheet for " + d.Name + " for " + dateStr}, -1)
	sheet.AddRow()
	sheet.AddRow().WriteSlice(&[]string{
		"Date",
		"Time",
		"Client Name",
		"Phone Number",
		"Pick Up",
		"Destination",
		"In Car Time",
		"Waiting Time (mins)",
		"Drop",
		"Miles",
		"Hours",
		"Parking (GBP)",
	}, -1)
	lastDate := time.Unix(0, 0)
	for n, ev := range e {
		var (
			ef EventFinals
			cl Client
		)
		if err := c.GetEventFinals(ev.ID, &ef); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if err := c.GetClient(ev.ClientID, &cl); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		record := make([]string, 6, 12)
		thisDate := time.Unix((ev.Start-(ev.Start%(24*3600000)))/1000, 0)
		record[0] = formatDate(ev.Start)
		if n > 0 {
			if !thisDate.Equal(lastDate) {
				sheet.AddRow()
			} else {
				record[0] = ""
			}
		}
		lastDate = thisDate
		record[1] = formatTime(ev.Start)
		record[2] = cl.Name
		record[3] = cl.PhoneNumber
		record[4] = ev.From
		record[5] = ev.To
		if ef.FinalsSet {
			record = append(record,
				formatTime(ef.InCar),
				strconv.Itoa(int(ef.Waiting)),
				formatTime(ef.Drop),
				strconv.Itoa(int(ef.Miles)),
				formatTime(ef.DriverHours),
				formatMoney(ef.Parking),
			)
		}
		sheet.AddRow().WriteSlice(&record, -1)
	}
	var buf []byte
	if err = xls.Write(memio.Create(&buf)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "inline; filename=\"driverEvents-"+d.Name+"-"+dateStr+".xlsx\"")
	http.ServeContent(w, r, "", time.Now(), memio.Open(buf))
}

func (c *Calls) exportClientEvents(w http.ResponseWriter, r *http.Request) {
	var f EventsFilter
	err := form.Parse(&f, r.PostForm)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if e, ok := err.(form.Errors); ok {
			for k, v := range e {
				fmt.Fprintf(w, "%s = %s\n", k, v)
			}
		} else {
			w.Write([]byte(err.Error()))
		}
		return
	}
	if f.Start > f.End {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid times"))
		return
	}
	var (
		e  []Event
		cl Client
		cy Company
	)
	dateStr := formatDate(f.Start)
	if f.Start != f.End {
		dateStr += " to " + formatDate(f.End)
	}
	f.End += 24 * 3600 * 1000
	err = c.ClientEvents(f, &e)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	err = c.GetClient(f.ID, &cl)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	err = c.GetCompany(cl.CompanyID, &cy)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	xls := xlsx.NewFile()
	sheet := xls.AddSheet("Client Events")
	sheet.AddRow().WriteSlice(&[]string{"Client Events for " + cl.Name + " for " + dateStr}, -1)
	sheet.AddRow()
	sheet.AddRow().WriteSlice(&[]string{
		"Driver",
		"From",
		"To",
		"Start",
		"End",
		"In Car",
		"Waiting",
		"Drop Off",
		"Trip Time",
		"Price (GBP)",
	}, -1)
	for _, ev := range e {
		var (
			ef EventFinals
			d  Driver
		)
		if err := c.GetEventFinals(ev.ID, &ef); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if err := c.GetDriver(ev.DriverID, &d); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		record := make([]string, 5, 10)
		record[0] = d.Name
		record[1] = ev.From
		record[2] = ev.To
		record[3] = formatDateTime(ev.Start)
		record[4] = formatDateTime(ev.End)
		if ef.FinalsSet {
			record = append(record,
				formatTime(ef.InCar),
				strconv.Itoa(int(ef.Waiting)),
				formatTime(ef.Drop),
				formatTime(ef.Trip),
				formatMoney(ef.Price),
			)
		}
		sheet.AddRow().WriteSlice(&record, -1)
	}
	var buf []byte
	if err = xls.Write(memio.Create(&buf)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "inline; filename=\"clientEvents-"+cl.Name+"-"+dateStr+".xlsx\"")
	http.ServeContent(w, r, "", time.Now(), memio.Open(buf))
}

func (c *Calls) exportCompanyEvents(w http.ResponseWriter, r *http.Request) {
	var f EventsFilter
	err := form.Parse(&f, r.PostForm)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if e, ok := err.(form.Errors); ok {
			for k, v := range e {
				fmt.Fprintf(w, "%s = %s\n", k, v)
			}
		} else {
			w.Write([]byte(err.Error()))
		}
		return
	}
	if f.Start > f.End {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid times"))
		return
	}
	dateStr := formatDate(f.Start)
	if f.Start != f.End {
		dateStr += " to " + formatDate(f.End)
	}
	f.End += 24 * 3600 * 1000
	var cy Company
	err = c.GetCompany(f.ID, &cy)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	var e []Event
	err = c.CompanyEvents(f, &e)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	xls := xlsx.NewFile()
	sheet := xls.AddSheet("Company Events")
	sheet.AddRow().WriteSlice(&[]string{"Company Events for " + cy.Name + " for " + dateStr}, -1)
	sheet.AddRow()
	sheet.AddRow().WriteSlice(&[]string{
		"Start",
		"End",
		"Client",
		"Reference",
		"From",
		"To",
		"Driver",
		"Parking (GBP)",
		"Price (GBP)",
	}, -1)
	for _, ev := range e {
		var (
			ef EventFinals
			cl Client
			d  Driver
		)
		if err := c.GetEventFinals(ev.ID, &ef); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if err := c.GetClient(ev.ClientID, &cl); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if err := c.GetDriver(ev.DriverID, &d); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		record := make([]string, 7, 9)
		record[0] = formatDateTime(ev.Start)
		record[1] = formatDateTime(ev.End)
		record[2] = cl.Name
		record[3] = cl.Reference
		record[4] = ev.From
		record[5] = ev.To
		record[6] = d.Name
		if ef.FinalsSet {
			record = append(record,
				formatMoney(ef.Parking),
				formatMoney(ef.Price),
			)
		}
		sheet.AddRow().WriteSlice(&record, -1)
	}
	var buf []byte
	if err = xls.Write(memio.Create(&buf)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "inline; filename=\"companyEvents-"+cy.Name+"-"+dateStr+".xlsx\"")
	http.ServeContent(w, r, "", time.Now(), memio.Open(buf))
}

func (c *Calls) exportCompanyClients(w http.ResponseWriter, r *http.Request) {
	var companyID int64
	err := form.ParseValue("id", form.Int64{&companyID}, r.PostForm)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if e, ok := err.(form.Errors); ok {
			for k, v := range e {
				fmt.Fprintf(w, "%s = %s\n", k, v)
			}
		} else {
			w.Write([]byte(err.Error()))
		}
		return
	}
	var cy Company
	err = c.GetCompany(companyID, &cy)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	var cl []Client
	err = c.ClientsForCompany(companyID, &cl)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	xls := xlsx.NewFile()
	sheet := xls.AddSheet("Client List for " + cy.Name)
	sheet.AddRow().WriteSlice(&[]string{"Client List for " + cy.Name}, -1)
	sheet.AddRow()
	sheet.AddRow().WriteSlice(&[]string{
		"Client Name",
		"Reference",
		"Phone Number",
	}, -1)
	for _, client := range cl {
		sheet.AddRow().WriteSlice(&[]string{
			client.Name,
			" " + client.Reference,
			client.PhoneNumber,
		}, -1)
	}
	var buf []byte
	if err = xls.Write(memio.Create(&buf)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "inline; filename=\"clientList-"+cy.Name+".xlsx\"")
	http.ServeContent(w, r, "", time.Now(), memio.Open(buf))
}

func (c *Calls) exportCompanyList(w http.ResponseWriter, r *http.Request) {
	var cy []Company
	err := c.Companies(struct{}{}, &cy)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if e, ok := err.(form.Errors); ok {
			for k, v := range e {
				fmt.Fprintf(w, "%s = %s\n", k, v)
			}
		} else {
			w.Write([]byte(err.Error()))
		}
		return
	}
	xls := xlsx.NewFile()
	sheet := xls.AddSheet("Company List")
	sheet.AddRow().WriteSlice(&[]string{"Company List"}, -1)
	sheet.AddRow()
	sheet.AddRow().WriteSlice(&[]string{
		"Company Name",
		"Address",
	}, -1)
	for _, company := range cy {
		sheet.AddRow().WriteSlice(&[]string{
			company.Name,
			company.Address,
		}, -1)
	}
	var buf []byte
	if err = xls.Write(memio.Create(&buf)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "inline; filename=\"companyList.xlsx\"")
	http.ServeContent(w, r, "", time.Now(), memio.Open(buf))
}

func (c *Calls) exportClientList(w http.ResponseWriter, r *http.Request) {
	var cl []Client
	err := c.Clients(struct{}{}, &cl)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	xls := xlsx.NewFile()
	sheet := xls.AddSheet("Client List")
	sheet.AddRow().WriteSlice(&[]string{"Client List"}, -1)
	sheet.AddRow()
	sheet.AddRow().WriteSlice(&[]string{
		"Client Name",
		"Company Name",
		"Reference",
		"Phone Number",
	}, -1)
	for _, client := range cl {
		var cy Company
		if err := c.GetCompany(client.CompanyID, &cy); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		sheet.AddRow().WriteSlice(&[]string{
			client.Name,
			cy.Name,
			client.Reference,
			" " + client.PhoneNumber,
		}, -1)
	}
	var buf []byte
	if err = xls.Write(memio.Create(&buf)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "inline; filename=\"clientList.xlsx\"")
	http.ServeContent(w, r, "", time.Now(), memio.Open(buf))
}
