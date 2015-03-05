package main

import (
	"errors"
	"net/rpc"
	"sync"
	"time"

	"github.com/MJKWoolnough/store"
)

var ErrInvalidPhoneNumber = errors.New("Invalid Phone Number Format")

type PhoneNumber uint64

func (p PhoneNumber) MarshalJSON() ([]byte, error) {
	var digits [21]byte
	pos := 21
	for num := uint64(p); num > 0; num /= 10 {
		pos--
		digits[pos] = '0' + byte(num%10)
	}
	if pos == 11 {
		pos--
		digits[pos] = '0'
	} else if pos == 10 && digits[10] == '4' && digits[11] == '4' {
		pos--
		digits[pos] = '+'
	} else {
		return digits[pos:], ErrInvalidPhoneNumber
	}
	return digits[pos:], nil
}

func (p *PhoneNumber) UnmarshalJSON(a []byte) error {
	var num uint64
	for _, digit := range a {
		if digit >= '0' && digit <= '9' {
			num *= 10
			num += uint64(digit - '0')
		}
	}
	*p = PhoneNumber(num)
	if (num < 7000000000 || num >= 8000000000) && (num < 447000000000 || num >= 448000000000) {
		return ErrInvalidPhoneNumber
	}
	return nil
}

type Driver struct {
	ID                 int64
	Name               string
	RegistrationNumber string
	PhoneNumber        PhoneNumber
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

func (c Calls) SetDriver(d Driver, id *int64) error {
	err := c.s.Set(&d)
	*id = d.ID
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
