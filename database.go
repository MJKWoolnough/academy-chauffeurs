package main

import (
	"time"

	"github.com/MJKWoolnough/form"
	"github.com/MJKWoolnough/store"
)

type Driver struct {
	ID                        int
	Name, Registration, Phone string
}

func (d *Driver) Get() store.TypeMap {
	return store.TypeMap{
		"id":           &d.ID,
		"name":         &d.Name,
		"registration": &d.Registration,
		"phone":        &d.Phone,
	}
}

func (d *Driver) ParserList() form.ParserList {
	return form.ParserList{
		"id":           form.Int{&d.ID},
		"name":         form.String{&d.Name},
		"registration": form.String{&d.Registration},
		"phone":        form.String{&d.Phone},
	}
}

func (Driver) Key() string {
	return "id"
}

func (Driver) TableName() string {
	return "drivers"
}

type Company struct {
	ID                       int
	Name, Address, Reference string
}

func (c *Company) Get() store.TypeMap {
	return store.TypeMap{
		"id":      &c.ID,
		"name":    &c.Name,
		"address": &c.Address,
		"ref":     &c.Reference,
	}
}

func (c *Company) ParserList() form.ParserList {
	return form.ParserList{
		"id":      form.Int{&c.ID},
		"name":    form.String{&c.Name},
		"address": form.String{&c.Address},
		"ref":     form.String{&c.Reference},
	}
}

func (Company) Key() string {
	return "id"
}

func (Company) TableName() string {
	return "companies"
}

type Client struct {
	ID, CompanyID                                int
	Name, CompanyName, Reference, Address, Phone string
}

func (c *Client) Get() store.TypeMap {
	return store.TypeMap{
		"id":        &c.ID,
		"companyID": &c.CompanyID,
		"name":      &c.Name,
		"ref":       &c.Reference,
		"address":   &c.Address,
		"phone":     &c.Phone,
	}
}

func (c *Client) ParserList() form.ParserList {
	return form.ParserList{
		"id":          form.Int{&c.ID},
		"companyName": form.String{&c.CompanyName},
		"name":        form.String{&c.Name},
		"ref":         form.String{&c.Reference},
		"address":     form.String{&c.Address},
		"phone":       form.String{&c.Phone},
	}
}

func (c *Client) companyName(db *store.Store) {
	comp := new(Company)
	comp.ID = c.CompanyID
	db.Get(comp)
	c.CompanyName = comp.Name
}

func (c *Client) companyID(db *store.Store) {
	comp := new(Company)
	db.Search([]store.Interface{comp}, 0, store.MatchString("name", c.CompanyName))
	c.CompanyID = comp.ID
}

func (Client) Key() string {
	return "id"
}

func (Client) TableName() string {
	return "clients"
}

type Event struct {
	ID        int
	Start     time.Time
	From, To  string
	EDuration int64
	Pickup    time.Time
	Waiting   int64
	Parking   int
	Miles     int
	ClientId  int
	Driver
}

func (e *Event) Get() store.TypeMap {
	return store.TypeMap{
		"id":                 &e.ID,
		"start":              &e.Start,
		"from":               &e.From,
		"to":                 &e.To,
		"estimatedDuration":  &e.EDuration,
		"pickupTime":         &e.Pickup,
		"waitingTime":        &e.Waiting,
		"parkingCosts":       &e.Parking,
		"milesDriver":        &e.Miles,
		"clientId":           &e.ClientId,
		"driverId":           &e.Driver.ID,
		"driverName":         &e.Driver.Name,
		"driverRegistration": &e.Driver.Registration,
		"driverPhone":        &e.Driver.Phone,
	}
}

func (e *Event) ParserList() form.ParserList {
	return form.ParserList{
		"id":                 form.Int{&e.ID},
		"start":              form.Time{&e.Start},
		"from":               form.String{&e.From},
		"to":                 form.String{&e.To},
		"estimatedDuration":  form.Int64{&e.EDuration},
		"pickupTime":         form.Time{&e.Pickup},
		"waitingTime":        form.Int64{&e.Waiting},
		"parkingCosts":       form.Int{&e.Parking},
		"milesDriver":        form.Int{&e.Miles},
		"clientId":           form.Int{&e.ClientId},
		"driverId":           form.Int{&e.Driver.ID},
		"driverName":         form.String{&e.Driver.Name},
		"driverRegistration": form.String{&e.Driver.Registration},
		"driverPhone":        form.String{&e.Driver.Phone},
	}
}

func (Event) Key() string {
	return "id"
}

func (Event) TableName() string {
	return "events"
}

func SetupDatabase(fname string) (*store.Store, error) {
	s, err := store.NewStore(fname)
	if err != nil {
		return nil, err
	}
	err = s.Register(new(Driver))
	if err != nil {
		return nil, err
	}
	err = s.Register(new(Company))
	if err != nil {
		return nil, err
	}
	err = s.Register(new(Client))
	if err != nil {
		return nil, err
	}
	err = s.Register(new(Event))
	if err != nil {
		return nil, err
	}
	return s, nil
}
