package main

import (
	"net/rpc"
	"sync"
	"time"

	"github.com/MJKWoolnough/store"
)

type Driver struct {
	ID                                    int64
	Name, RegistrationNumber, PhoneNumber string
}

type Company struct {
	ID   int64
	Name string
}

type Client struct {
	ID      int64
	Name    string
	Company Company
}

type Event struct {
	ID         int64
	Start, End time.Time
	Driver     Driver
	Client     Client
}

type search struct {
	ps     *store.PreparedSearch
	params interface{}
}

type Calls struct {
	s        *store.Store
	searches map[string]search
}

func newCalls(dbFName string) (*Calls, error) {
	s, err := store.New(dbFName)
	if err != nil {
		return nil, err
	}
	err = s.Register(new(Event))
	if err != nil {
		return nil, err
	}
	c := &Calls{
		s,
		make(map[string]search),
	}
	err = rpc.Register(c)
	if err != nil {
		return nil, err
	}
	// setup searches

	ef := new(eventsFilter)
	ns := s.NewSearch(new(Event))
	ns.Sort = []store.SortBy{{Column: "Start", Asc: true}}
	ns.Filter = store.And{
		idEqual{"ID", &ef.DriverID},
		store.Or{
			betweenTime{"Start", &ef.From, &ef.To},
			betweenTime{"End", &ef.From, &ef.To},
		},
	}
	ps, err := ns.Prepare()
	if err != nil {
		return nil, err
	}
	c.searches["events"] = search{ps, ef}

	return c, nil
}

func (c Calls) close() {
	c.s.Close()
}

type eventsFilter struct {
	DriverID int64
	From, To time.Time
	mu       sync.Mutex
}

func (c Calls) Events(f eventsFilter, eventList *[]Event) error {
	s := c.searches["events"]
	n, err := s.ps.Count()
	if err != nil {
		return err
	}
	e := s.params.(*eventsFilter)
	e.mu.Lock()
	defer e.mu.Unlock()
	e.DriverID = f.DriverID
	e.From = f.From
	e.To = f.To
	es := make([]Event, n)
	ei := make([]interface{}, 0, n)
	for i := 0; i < n; i++ {
		ei = append(ei, &es[i])
	}
	_, err = s.ps.GetPage(ei, 0)
	if err != nil {
		return err
	}
	*eventList = es
	return nil
}

func (c Calls) Drivers(_ byte, drivers *[]Driver) error {
	n, err := c.s.Count(new(Driver))
	if err != nil {
		return err
	}
	d := make([]Driver, n)
	di := make([]interface{}, 0, n)
	for i := 0; i < n; i++ {
		di = append(di, &d)
	}
	_, err = c.s.GetPage(di, 0)
	if err != nil {
		return err
	}
	*drivers = d
	return nil
}

type SetDriverResponse struct {
	Errors                          bool
	ID                              int64
	NameError, RegError, PhoneError string
}

func (c Calls) SetDriver(d Driver, resp *SetDriverResponse) error {
	if d.Name == "" {
		resp.Errors = true
		resp.NameError = "Name Required"
	}
	if d.RegistrationNumber == "" {
		resp.Errors = true
		resp.RegError = "Registration Number Required"
	}
	if !ValidMobileNumber(d.PhoneNumber) {
		resp.Errors = true
		resp.PhoneError = "Valid Mobile Phone Number Required"
	}
	var err error
	if !resp.Errors {
		err = c.s.Set(&d)
		resp.ID = d.ID
	}
	return err
}

// Filters

type betweenTime struct {
	col      string
	from, to *time.Time
}

func (b betweenTime) SQL() string {
	return "[" + b.col + "] BETWEEN ? AND ?"
}

func (b betweenTime) Vars() []interface{} {
	return []interface{}{b.from, b.to}
}

type idEqual struct {
	col string
	id  *int64
}

func (i idEqual) SQL() string {
	return "[" + i.col + "] = ?"
}

func (i idEqual) Vars() []interface{} {
	return []interface{}{i.id}
}
