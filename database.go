package main

import (
	"sync"

	"code.google.com/p/go-sqlite/go1/sqlite3"
)

type DB struct {
	*sqlite3.Conn
	addDriver    *sqlite3.Stmt
	getDriver    *sqlite3.Stmt
	getDrivers   *sqlite3.Stmt
	editDriver   *sqlite3.Stmt
	removeDriver *sqlite3.Stmt
	sync.Mutex
}

func NewDB(filename string) (*DB, error) {
	s, err := sqlite3.Open(filename)
	if err != nil {
		return nil, err
	}
	d := &DB{
		Conn: s,
	}
	err = d.prepareDriverStatements()
	if err != nil {
		return nil, err
	}
	return d, nil
}
