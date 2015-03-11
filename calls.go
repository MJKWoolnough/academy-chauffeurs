package main

import (
	"database/sql"
	"net/rpc"
	"strconv"
	"sync"
	"time"

	_ "github.com/mxk/go-sqlite/sqlite3"
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

type Time struct {
	time.Time
}

func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(t.Unix(), 10)), nil
}

func (t Time) UnmarshalJSON(data []byte) error {
	num, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	t.Time = time.Unix(num, 0)
	return nil
}

type Event struct {
	ID, DriverID, ClientID int64
	Start, End             Time
	From, To               string
}

const (
	CreateDriver = iota
	CreateCompany
	CreateClient
	CreateEvent

	ReadDriver
	ReadCompany
	ReadClient
	ReadEvent

	UpdateDriver
	UpdateCompany
	UpdateClient
	UpdateEvent

	DeleteDriver
	DeleteCompany
	DeleteClient
	DeleteEvent

	DriverList
	EventList
	EventOverlap

	TotalStmts
)

type Calls struct {
	mu         sync.Mutex
	db         *sql.DB
	statements [TotalStmts]*sql.Stmt
}

func newCalls(dbFName string) (*Calls, error) {
	db, err := sql.Open("sqlite3", dbFName)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return nil, err
	}

	// Tables

	for _, ct := range []string{
		"[Driver]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Name] TEXT, [RegistrationNumber] TEXT, [PhoneNumber] TEXT);",
		"[Company]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Name] TEXT, [Address] TEXT);",
		"[Client]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [CompanyID] INTEGER REFERENCES [Company]([ID]) ON DELETE CASCADE, [Name] TEXT, [PhoneNumber] TEXT, [Reference] TEXT);",
		"[Event]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [DriverID] INTEGER REFERENCES [Driver]([ID]) ON DELETE CASCADE, [ClientID] INTEGER REFERENCES [Client]([ID]) ON DELETE CASCADE, [Start] INTEGER, [End] INTEGER, [From] TEXT, [To] TEXT);",
	} {
		if _, err = db.Exec("CREATE TABLE IF NOT EXISTS " + ct); err != nil {
			return nil, err
		}
	}

	c := &Calls{
		db: db,
	}

	for n, ps := range []string{
		// Create

		"INSERT INTO [Driver]([Name], [RegistrationNumber], [PhoneNumber]) VALUES (?, ?, ?);",
		"INSERT INTO [Company]([Name], [Address]) VALUES (?, ?);",
		"INSERT INTO [Client]([CompanyID], [Name], [PhoneNumber], [Reference]) VALUES (?, ?, ?, ?);",
		"INSERT INTO [Event]([DriverID], [ClientID], [Start], [End], [From], [To]) VALUES (?, ?, ?, ?, ?, ?);",

		// Read

		"SELECT [Name], [RegistrationNumber], [PhoneNumber] FROM [Driver] WHERE [ID] = ?;",
		"SELECT [Name], [Address] FROM [Company] WHERE [ID] = ?;",
		"SELECT [CompanyID], [Name], [PhoneNumber], [Reference] FROM [Client] WHERE [ID] = ?;",
		"SELECT [DriverID], [ClientID], [Start], [End], [From], [To] FROM [Event] WHERE [ID] = ?;",

		// Update

		"UPDATE [Driver] SET [Name] = ?, [RegistrationNumber] = ?, [PhoneNumber] = ? WHERE [ID] = ?;",
		"UPDATE [Company] SET [Name] = ?, [Address] = ? WHERE [ID] = ?;",
		"UPDATE [Client] SET [CompanyID] = ?, [Name] = ?, [PhoneNumber] = ?, [Reference] = ? WHERE [ID] = ?;",
		"UPDATE [Event] SET [DriverID] = ?, [ClientID] = ?, [Start] = ?, [End] = ?, [From] = ?, [To] = ? WHERE [ID] = ?;",

		// Delete

		"DELETE FROM [Driver] WHERE [ID] = ?;",
		"DELETE FROM [Company] WHERE [ID] = ?;",
		"DELETE FROM [Client] WHERE [ID] = ?;",
		"DELETE FROM [Event] WHERE [ID] = ?;",

		// Searches

		// All Drivers
		"SELECT [ID], [Name], [RegistrationNumber], [PhoneNumber] FROM [Driver];",
		// Row of Events for driver
		"SELECT [DriverID], [ClientID], [Start], [End], [From], [To] FROM [Event] WHERE [DriverID] = ? AND ([Start] Between ? AND ? OR [End] Between ?2 AND ?3);",
		// Event Overlaps
		"SELECT COUNT(1) FROM [Event] WHERE [ID] != ? AND [DriverID] = ? AND ([Start] Between ? AND ? OR [End] Between ?3 AND ?4);",
	} {
		stmt, err := db.Prepare(ps)
		if err != nil {
			return nil, err
		}
		c.statements[n] = stmt
	}

	err = rpc.Register(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Calls) close() {
	c.db.Close()
}

type EventsFilter struct {
	DriverID   int64
	Start, End time.Time
	mu         sync.Mutex
}

func (c *Calls) Events(f EventsFilter, eventList *[]Event) error {
	rows, err := c.statements[EventList].Query(f.DriverID, f.Start, f.End)
	if err != nil {
		return err
	}
	for rows.Next() {
		var e Event
		err = rows.Scan(&e.DriverID, &e.ClientID, &e.Start, &e.End, &e.From, &e.To)
		if err != nil {
			return err
		}
		(*eventList) = append(*eventList, e)
	}
	return rows.Err()
}

func (c *Calls) Drivers(_ struct{}, drivers *[]Driver) error {
	rows, err := c.statements[DriverList].Query()
	if err != nil {
		return err
	}
	for rows.Next() {
		var d Driver
		err = rows.Scan(&d.ID, &d.Name, &d.RegistrationNumber, &d.PhoneNumber)
		if err != nil {
			return err
		}
		(*drivers) = append(*drivers, d)
	}
	return rows.Err()
}
