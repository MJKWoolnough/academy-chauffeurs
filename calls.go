package main

import (
	"net/rpc"

	"github.com/MJKWoolnough/store"
)

type Driver struct {
	ID   int64
	Name string
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
	ID     int64
	Driver Driver
	Client Client
}

type Calls struct {
	s *store.Store
}

func newCalls(dbFName string) (Calls, error) {
	s, err := store.New(dbFName)
	if err != nil {
		return calls{}, err
	}
	err = s.Register(new(Event))
	if err != nil {
		return calls{}, err
	}
	c := calls{s}
	err = rpc.Register(c)
	if err != nil {
		return calls{}, err
	}
	return c, nil
}

func (c Calls) close() {
	c.s.Close()
}

func (c Calls) Test(testID *int, response *int) error {
	*response = *testID + 1
	return nil
}
