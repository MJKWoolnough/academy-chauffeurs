package main

import (
	"database/sql"
	"net/rpc"
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

type Event struct {
	ID, DriverID, ClientID int64
	Start, End             int64
	From, To               string
	MessageSent            bool
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
	DriverEvents
	EventOverlap
	CompanyList
	ClientList
	UnsentMessages

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
		"[Event]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [DriverID] INTEGER REFERENCES [Driver]([ID]) ON DELETE CASCADE, [ClientID] INTEGER REFERENCES [Client]([ID]) ON DELETE CASCADE, [Start] INTEGER, [End] INTEGER, [From] TEXT, [To] TEXT, [MessageSent] BOOLEAN DEFAULT 0 NOT NULL CHECK ([MessageSent] IN (0,1)));",
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
		"SELECT [ID], [Name], [RegistrationNumber], [PhoneNumber] FROM [Driver] ORDER BY [ID] ASC;",

		// Row of Events for driver
		"SELECT [ID], [DriverID], [ClientID], [Start], [End], [From], [To] FROM [Event] WHERE [DriverID] = ? AND [Start] Between ? AND ?;",
		// Event Overlaps
		"SELECT COUNT(1) FROM [Event] WHERE [ID] != ? AND [DriverID] = ? AND ([Start] Between ? AND ? OR [End] Between ?3 AND ?4);",

		// Company List
		"SELECT [ID], [Name], [Address] FROM [Company] ORDER BY [ID] ASC;",

		// Client List
		"SELECT [ID], [CompanyID], [Name], [PhoneNumber], [Reference] FROM [Client] ORDER BY [ID] ASC;",

		// Events with unsent messages
		"SELECT [ID], [DriverID], [ClientID], [Start], [End], [From], [To] FROM [Event] WHERE [MessageSent] = 0 AND [Start] > ?;",
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

type is []interface{}

func (c *Calls) getList(sqlStmt int, params is, get func() is) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	rows, err := c.statements[sqlStmt].Query(params...)
	if err != nil {
		return err
	}
	for rows.Next() {
		if err = rows.Scan(get()...); err != nil {
			return err
		}
	}
	return rows.Err()
}

type EventsFilter struct {
	DriverID, Start, End int64
}

func (c *Calls) DriverEvents(f EventsFilter, events *[]Event) error {
	*events = make([]Event, 0)
	return c.getList(DriverEvents, is{f.DriverID, f.Start, f.End}, func() is {
		var (
			e   Event
			pos = len(*events)
		)
		*events = append(*events, e)
		return is{
			&(*events)[pos].ID,
			&(*events)[pos].DriverID,
			&(*events)[pos].ClientID,
			&(*events)[pos].Start,
			&(*events)[pos].End,
			&(*events)[pos].From,
			&(*events)[pos].To,
		}
	})
}

func (c *Calls) Drivers(_ struct{}, drivers *[]Driver) error {
	*drivers = make([]Driver, 0)
	return c.getList(DriverList, is{}, func() is {
		var (
			d   Driver
			pos = len(*drivers)
		)
		*drivers = append(*drivers, d)
		return is{
			&(*drivers)[pos].ID,
			&(*drivers)[pos].Name,
			&(*drivers)[pos].RegistrationNumber,
			&(*drivers)[pos].PhoneNumber,
		}
	})
}

func (c *Calls) Companies(_ struct{}, companies *[]Company) error {
	*companies = make([]Company, 0)
	return c.getList(CompanyList, is{}, func() is {
		var (
			cy  Company
			pos = len(*companies)
		)
		*companies = append(*companies, cy)
		return is{
			&(*companies)[pos].ID,
			&(*companies)[pos].Name,
			&(*companies)[pos].Address,
		}
	})
}

func (c *Calls) Clients(_ struct{}, clients *[]Client) error {
	*clients = make([]Client, 0)
	return c.getList(ClientList, is{}, func() is {
		var (
			cl  Client
			pos = len(*clients)
		)
		*clients = append(*clients, cl)
		return is{
			&(*clients)[pos].ID,
			&(*clients)[pos].CompanyID,
			&(*clients)[pos].Name,
			&(*clients)[pos].PhoneNumber,
			&(*clients)[pos].Reference,
		}
	})
}

func (c *Calls) UnsentMessages(_ struct{}, events *[]Event) error {
	*events = make([]Event, 0)
	return c.getList(UnsentMessages, is{time.Now()}, func() is {
		var (
			e   Event
			pos = len(*events)
		)
		*events = append(*events, e)
		return is{
			&(*events)[pos].ID,
			&(*events)[pos].DriverID,
			&(*events)[pos].ClientID,
			&(*events)[pos].Start,
			&(*events)[pos].End,
			&(*events)[pos].From,
			&(*events)[pos].To,
		}
	})
}

type Message struct {
}

func (c *Calls) SendMessage(m Message, _ *struct{}) error {
	return nil
}

type AutocompleteAddressRequest struct {
	Priority int
	Partial  string
}

type AutocompleteValue struct {
	ID    int64
	Value string
}

func makeAutocompleteSQL(column, table string, notIDs []int64) string {
	sql := "SELECT [ID], [" + column + "] FROM [" + table + "] WHERE [" + column + "] LIKE ?"
	if len(notIDs) > 0 {
		sql += " AND [ID] NOT IN ("
		for n := range notIDs {
			if n > 0 {
				sql += ", "
			}
			sql += "?"
		}
		sql += ")"
	}
	return sql + " LIMIT ?;"
}

const MAXRETURN = 10

func (c *Calls) autocomplete(vals *[]AutocompleteValue, column, table, partial string, notIDs ...int64) error {
	if len(*vals) >= MAXRETURN {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	stmt, err := c.db.Prepare(makeAutocompleteSQL(column, table, notIDs))
	if err != nil {
		return err
	}
	params := make([]interface{}, 1, len(notIDs)+2)
	params[0] = partial
	for _, n := range notIDs {
		params = append(params, n)
	}
	params = append(params, MAXRETURN-len(notIDs))
	rows, err := stmt.Query(params...)
	if err != nil {
		return err
	}
	for rows.Next() {
		var a AutocompleteValue
		err := rows.Scan(&a.ID, &a.Value)
		if err != nil {
			return err
		}
		*vals = append(*vals, a)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return stmt.Close()
}

func (c *Calls) AutocompleteCompanyName(partial string, vals *[]AutocompleteValue) error {
	*vals = make([]AutocompleteValue, 0, MAXRETURN)
	err := c.autocomplete(vals, "Name", "Company", partial+"%")
	if err != nil || len(*vals) >= MAXRETURN {
		return err
	}
	notIDs := make([]int64, 0, MAXRETURN)
	for _, v := range *vals {
		notIDs = append(notIDs, v.ID)
	}
	return c.autocomplete(vals, "Name", "Company", "%"+partial+"%", notIDs...)
}

func (c *Calls) AutocompleteClientName(partial string, vals *[]AutocompleteValue) error {
	*vals = make([]AutocompleteValue, 0, MAXRETURN)
	err := c.autocomplete(vals, "Name", "Client", partial+"%")
	if err != nil || len(*vals) >= MAXRETURN {
		return err
	}
	notIDs := make([]int64, 0, MAXRETURN)
	for _, v := range *vals {
		notIDs = append(notIDs, v.ID)
	}
	return c.autocomplete(vals, "Name", "Client", "%"+partial+"%", notIDs...)
}

func (c *Calls) AutocompleteAddress(req AutocompleteAddressRequest, vals *[]AutocompleteValue) error {
	*vals = make([]AutocompleteValue, 0, MAXRETURN)
	var first, second string
	if req.Priority == 0 {
		first = "From"
		second = "To"
	} else {
		first = "From"
		second = "To"
	}
	err := c.autocomplete(vals, first, "Event", req.Partial+"%")
	if err != nil || len(*vals) >= MAXRETURN {
		return err
	}
	notIDs := make([]int64, 0, MAXRETURN)
	for _, v := range *vals {
		notIDs = append(notIDs, v.ID)
	}
	err = c.autocomplete(vals, second, "Event", req.Partial+"%", notIDs...)
	if err != nil || len(*vals) >= MAXRETURN {
		return err
	}
	for _, v := range *vals {
		notIDs = append(notIDs, v.ID)
	}
	err = c.autocomplete(vals, first, "Event", "%"+req.Partial+"%", notIDs...)
	if err != nil || len(*vals) >= MAXRETURN {
		return err
	}
	for _, v := range *vals {
		notIDs = append(notIDs, v.ID)
	}
	return c.autocomplete(vals, second, "Event", "%"+req.Partial+"%", notIDs...)
}
