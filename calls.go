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
	ID            int64
	Name, Address string
}

type Client struct {
	ID, CompanyID                int64
	Name, PhoneNumber, Reference string
}

type Event struct {
	ID, DriverID, ClientID int64
	Start, End             time.Time
	From, To               string
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
	err = s.Register(new(Driver), new(Company), new(Client), new(Event))
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

	ef := new(EventsFilter)
	ns := s.NewSearch(new(Event))
	ns.Sort = []store.SortBy{{Column: "Start", Asc: true}}
	ns.Filter = store.And{
		idNotEqual{"ID", &ef.NotID},
		idEqual{"DriverID", &ef.DriverID},
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

type EventsFilter struct {
	NotID, DriverID int64
	From, To        time.Time
	mu              sync.Mutex
}

func (c Calls) Events(f EventsFilter, eventList *[]Event) error {
	s := c.searches["events"]
	n, err := s.ps.Count()
	if err != nil {
		return err
	}
	e := s.params.(*EventsFilter)
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
		di = append(di, &d[i])
	}
	_, err = c.s.GetPage(di, 0)
	if err != nil {
		return err
	}
	*drivers = d
	return nil
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

type idNotEqual struct {
	col string
	id  *int64
}

func (i idNotEqual) SQL() string {
	return "[" + i.col + "] != ?"
}

func (i idNotEqual) Vars() []interface{} {
	return []interface{}{i.id}
}
