package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/MJKWoolnough/form"
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
	dateStr := formatDate(f.Start)
	if f.Start != f.End {
		dateStr += " to " + formatDate(f.End)
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "inline; filename=\"driverEvents-"+d.Name+"-"+dateStr+".csv\"")
	cw := csv.NewWriter(w)
	cw.Write([]string{"Driver Sheet for " + d.Name + " for " + dateStr})
	cw.Write([]string{})
	cw.Write([]string{
		"Start",
		"End",
		"Client Name",
		"Phone Number",
		"From",
		"To",
		"Company Name",
		"In Car Time",
		"Waiting Time (mins)",
		"Miles",
		"Trip Time",
		"Driver Hours",
		"Parking (GBP)",
	})
	for _, ev := range e {
		var (
			ef EventFinals
			cl Client
			cy Company
		)
		c.GetEventFinals(ev.ID, &ef)
		c.GetClient(ev.ClientID, &cl)
		c.GetCompany(cl.CompanyID, &cy)
		record := make([]string, 7, 13)
		record[0] = formatDateTime(ev.Start)
		record[1] = formatDateTime(ev.End)
		record[2] = cl.Name
		record[3] = " " + cl.PhoneNumber
		record[4] = ev.From
		record[5] = ev.To
		record[6] = cy.Name
		if ef.FinalsSet {
			record = append(record,
				formatTime(ef.InCar),
				strconv.Itoa(int(ef.Waiting)),
				strconv.Itoa(int(ef.Miles)),
				formatTime(ef.Trip),
				formatTime(ef.DriverHours),
				formatMoney(ef.Parking),
			)
		}
		cw.Write(record)
	}
	cw.Flush()
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
	dateStr := formatDate(f.Start)
	if f.Start != f.End {
		dateStr += " to " + formatDate(f.End)
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "inline; filename=\"clientEvents-"+cl.Name+"-"+dateStr+".csv\"")
	cw := csv.NewWriter(w)
	cw.Write([]string{"Client Events for " + cl.Name + " for " + dateStr})
	cw.Write([]string{})
	cw.Write([]string{
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
	})
	for _, ev := range e {
		var (
			ef EventFinals
			d  Driver
		)
		c.GetEventFinals(ev.ID, &ef)
		c.GetDriver(ev.DriverID, &d)
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
		cw.Write(record)
	}
	cw.Flush()
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
	dateStr := formatDate(f.Start)
	if f.Start != f.End {
		dateStr += " to " + formatDate(f.End)
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "inline; filename=\"companyEvents-"+cy.Name+"-"+dateStr+".csv\"")
	cw := csv.NewWriter(w)
	cw.Write([]string{"Company Events for " + cy.Name + " for " + dateStr})
	cw.Write([]string{})
	cw.Write([]string{
		"Start",
		"End",
		"Client",
		"Reference",
		"From",
		"To",
		"Driver",
		"Parking (GBP)",
		"Price (GBP)",
	})
	for _, ev := range e {
		var (
			ef EventFinals
			cl Client
			d  Driver
		)
		c.GetEventFinals(ev.ID, &ef)
		c.GetClient(ev.ClientID, &cl)
		c.GetDriver(ev.DriverID, &d)
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
		cw.Write(record)
	}
	cw.Flush()
}

func (c *Calls) exportCompanyClients(w http.ResponseWriter, r *http.Request) {
	var companyID int64
	err := form.ParseValue("companyID", form.Int64{&companyID}, r.PostForm)
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
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "inline; filename=\"clientList-"+cy.Name+".csv\"")
	cw := csv.NewWriter(w)
	cw.Write([]string{"Client List for " + cy.Name})
	cw.Write([]string{})
	cw.Write([]string{
		"Client Name",
		"Reference",
		"Phone Number",
	})
	for _, client := range cl {
		cw.Write([]string{
			client.Name,
			" " + client.Reference,
			client.PhoneNumber,
		})
	}
	cw.Flush()
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
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "inline; filename=\"companyList.csv\"")
	cw := csv.NewWriter(w)
	cw.Write([]string{"Company List"})
	cw.Write([]string{})
	cw.Write([]string{
		"Company Name",
		"Address",
	})
	for _, company := range cy {
		cw.Write([]string{
			company.Name,
			company.Address,
		})
	}
	cw.Flush()
}

func (c *Calls) exportClientList(w http.ResponseWriter, r *http.Request) {
	var cl []Client
	err := c.Clients(struct{}{}, &cl)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "inline; filename=\"clientList.csv\"")
	cw := csv.NewWriter(w)
	cw.Write([]string{"Client List"})
	cw.Write([]string{})
	cw.Write([]string{
		"Client Name",
		"Company Name",
		"Reference",
		"Phone Number",
	})
	for _, client := range cl {
		var cy Company
		c.GetCompany(client.CompanyID, &cy)
		cw.Write([]string{
			client.Name,
			cy.Name,
			client.Reference,
			" " + client.PhoneNumber,
		})
	}
	cw.Flush()
}
