package main

import "github.com/MJKWoolnough/store"

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
