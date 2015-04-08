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

	DeleteDriverEvents
	DeleteClientEvents
	DeleteCompanyClients
	DeleteCompanyEvents

	DriverList
	DriverEvents
	ClientEvents
	CompanyEvents
	EventOverlap
	CompanyList
	ClientList
	ClientForCompanyList
	UnsentMessages
	DisambiguateClients

	NumClientsForCompany
	NumEventsForCompany
	NumEventsForClient

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
		"[Driver]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Name] TEXT, [RegistrationNumber] TEXT, [PhoneNumber] TEXT, [Deleted] BOOLEAN DEFAULT 0 NOT NULL CHECK ([Deleted] IN (0,1)));",
		"[Company]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Name] TEXT, [Address] TEXT, [Deleted] BOOLEAN DEFAULT 0 NOT NULL CHECK ([Deleted] IN (0,1)));",
		"[Client]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [CompanyID] INTEGER REFERENCES [Company]([ID]) ON DELETE CASCADE, [Name] TEXT, [PhoneNumber] TEXT, [Reference] TEXT, [Deleted] BOOLEAN DEFAULT 0 NOT NULL CHECK ([Deleted] IN (0,1)));",
		"[Event]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [DriverID] INTEGER REFERENCES [Driver]([ID]) ON DELETE CASCADE, [ClientID] INTEGER REFERENCES [Client]([ID]) ON DELETE CASCADE, [Start] INTEGER, [End] INTEGER, [From] TEXT, [To] TEXT, [InCar] INTEGER DEFAULT 0, [Parking] INTEGER DEFAULT 0, [Waiting] INTEGER DEFAULT 0, [Drop] INTEGER DEFAULT 0, [Miles] INTEGER DEFAULT 0, [Hours] INTEGER DEFAULT 0, [Other] TEXT, [MessageSent] BOOLEAN DEFAULT 0 NOT NULL CHECK ([MessageSent] IN (0,1)), [Deleted] BOOLEAN DEFAULT 0 NOT NULL CHECK ([Deleted] IN (0,1)));",
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

		"SELECT [Name], [RegistrationNumber], [PhoneNumber] FROM [Driver] WHERE [ID] = ? AND [Deleted] = 0;",
		"SELECT [Name], [Address] FROM [Company] WHERE [ID] = ? AND [Deleted] = 0;",
		"SELECT [CompanyID], [Name], [PhoneNumber], [Reference] FROM [Client] WHERE [ID] = ? AND [Deleted] = 0;",
		"SELECT [DriverID], [ClientID], [Start], [End], [From], [To] FROM [Event] WHERE [ID] = ? AND [Deleted] = 0;",

		// Update

		"UPDATE [Driver] SET [Name] = ?, [RegistrationNumber] = ?, [PhoneNumber] = ? WHERE [ID] = ?;",
		"UPDATE [Company] SET [Name] = ?, [Address] = ? WHERE [ID] = ?;",
		"UPDATE [Client] SET [CompanyID] = ?, [Name] = ?, [PhoneNumber] = ?, [Reference] = ? WHERE [ID] = ?;",
		"UPDATE [Event] SET [DriverID] = ?, [ClientID] = ?, [Start] = ?, [End] = ?, [From] = ?, [To] = ? WHERE [ID] = ?;",

		// Delete (set deleted)

		"UPDATE [Driver] SET [Deleted] = 1 WHERE [ID] = ?;",
		"UPDATE [Company] SET [Deleted] = 1 WHERE [ID] = ?;",
		"UPDATE [Client] SET [Deleted] = 1 WHERE [ID] = ?;",
		"UPDATE [Event] SET [Deleted] = 1 WHERE [ID] = ?;",

		"UPDATE [Event] SET [Deleted] = 1 WHERE [DriverID] = ?;",
		"UPDATE [Event] SET [Deleted] = 1 WHERE [ClientID] = ?;",
		"UPDATE [Client] SET [Deleted] = 1 WHERE [CompanyID] = ?;",
		"UPDATE [Event] SET [Deleted] = 1 WHERE [ClientID] IN (SELECT [ID] FROM [Client] WHERE [CompanyID] = ?);",

		// Searches

		// All Drivers
		"SELECT [ID], [Name], [RegistrationNumber], [PhoneNumber] FROM [Driver] WHERE [Deleted] = 0 ORDER BY [ID] ASC;",

		// Row of Events for driver
		"SELECT [ID], [DriverID], [ClientID], [Start], [End], [From], [To] FROM [Event] WHERE [DriverID] = ? AND [Deleted] = 0 AND [Start] Between ? AND ? ORDER BY [Start] ASC;",

		// Row of Events for client
		"SELECT [ID], [DriverID], [ClientID], [Start], [End], [From], [To] FROM [Event] WHERE [ClientID] = ? AND [Deleted] = 0 AND [Start] Between ? AND ? ORDER BY [Start] ASC;",

		// Row of Events for company
		"SELECT [ID], [DriverID], [ClientID], [Start], [End], [From], [To] FROM [Event] WHERE [ClientID] IN (SELECT [ID] FROM [Client] WHERE [CompanyID] = ?) AND [Deleted] = 0 AND [Start] Between ? AND ? ORDER BY [Start] ASC;",

		// Event Overlaps
		"SELECT COUNT(1) FROM [Event] WHERE [ID] != ? AND [Deleted] = 0 AND [DriverID] = ? AND ([Start] <= ? AND [START] > ? OR [End] <= ?3 AND [END] > ?4);",

		// Company List
		"SELECT [ID], [Name], [Address] FROM [Company] WHERE [Deleted] = 0 ORDER BY [ID] ASC;",

		// Client List
		"SELECT [ID], [CompanyID], [Name], [PhoneNumber], [Reference] FROM [Client] WHERE [Deleted] = 0 ORDER BY [ID] ASC;",

		// Clients for company
		"SELECT [ID], [CompanyID], [Name], [PhoneNumber], [Reference] FROM [Client] WHERE [CompanyID] = ? AND [Deleted] = 0 ORDER BY [ID] ASC;",

		// Events with unsent messages
		"SELECT [ID], [DriverID], [ClientID], [Start], [End], [From], [To] FROM [Event] WHERE [MessageSent] = 0 AND [Start] > ? AND [Deleted] = 0;",

		// Disambiguate clients
		"SELECT [Company].[Name], [Client].[Reference] FROM [Client] JOIN [Company] ON [Client].[CompanyID] = [Company].[ID] WHERE [Client].[ID] = ?;",

		// Num Clients for company
		"SELECT COUNT(1) FROM [Client] WHERE [CompanyID] = ? AND [Deleted] = 0;",

		// Num Events for company
		"SELECT COUNT(1) FROM [Event] JOIN [Client] ON [Event].[ClientID] = [Client].[ID] WHERE [Client].[CompanyID] = ? AND [Client].[Deleted] = 0 AND [Event].[Deleted] = 0;",

		// Num Events for client
		"SELECT COUNT(1) FROM [Event] WHERE [ClientID] = ? AND [Deleted] = 0;",
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

func (c *Calls) NumClients(id int64, num *int64) error {
	return c.statements[NumClientsForCompany].QueryRow(id).Scan(num)
}

func (c *Calls) NumEvents(id int64, num *int64) error {
	return c.statements[NumEventsForCompany].QueryRow(id).Scan(num)
}

func (c *Calls) NumEventsClient(id int64, num *int64) error {
	return c.statements[NumEventsForClient].QueryRow(id).Scan(num)
}

type EventsFilter struct {
	ID, Start, End int64
}

func (c *Calls) DriverEvents(f EventsFilter, events *[]Event) error {
	*events = make([]Event, 0)
	return c.getList(DriverEvents, is{f.ID, f.Start, f.End}, func() is {
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

func (c *Calls) ClientEvents(f EventsFilter, events *[]Event) error {
	*events = make([]Event, 0)
	return c.getList(ClientEvents, is{f.ID, f.Start, f.End}, func() is {
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

func (c *Calls) CompanyEvents(f EventsFilter, events *[]Event) error {
	*events = make([]Event, 0)
	return c.getList(CompanyEvents, is{f.ID, f.Start, f.End}, func() is {
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

func (c *Calls) ClientsForCompany(companyID int64, clients *[]Client) error {
	*clients = make([]Client, 0)
	return c.getList(ClientForCompanyList, is{companyID}, func() is {
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
	ID                    int64
	Value, Disambiguation string
}

func makeAutocompleteSQL(column, table string, includeDeleted bool, notIDs []int64) string {
	sql := "SELECT [ID], [" + column + "] FROM [" + table + "] WHERE [" + column + "] LIKE ?"
	if !includeDeleted {
		sql += " AND [Deleted] = 0"
	}
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

func (c *Calls) autocomplete(vals *[]AutocompleteValue, column, table, partial string, includeDeleted bool, notIDs ...int64) error {
	if len(*vals) >= MAXRETURN {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	stmt, err := c.db.Prepare(makeAutocompleteSQL(column, table, includeDeleted, notIDs))
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
	err := c.autocomplete(vals, "Name", "Company", partial+"%", false)
	if err != nil || len(*vals) >= MAXRETURN {
		return err
	}
	notIDs := make([]int64, 0, MAXRETURN)
	for _, v := range *vals {
		notIDs = append(notIDs, v.ID)
	}
	return c.autocomplete(vals, "Name", "Company", "%"+partial+"%", false, notIDs...)
}

func (c *Calls) AutocompleteClientName(partial string, vals *[]AutocompleteValue) error {
	*vals = make([]AutocompleteValue, 0, MAXRETURN)
	err := c.autocomplete(vals, "Name", "Client", partial+"%", false)
	if err != nil || len(*vals) >= MAXRETURN {
		return err
	}
	notIDs := make([]int64, 0, MAXRETURN)
	for _, v := range *vals {
		notIDs = append(notIDs, v.ID)
	}
	err = c.autocomplete(vals, "Name", "Client", "%"+partial+"%", false, notIDs...)
	if err != nil {
		return err
	}
	for n := range *vals {
		var cName, ref string
		err := c.statements[DisambiguateClients].QueryRow((*vals)[n].ID).Scan(&cName, &ref)
		if err != nil {
			return err
		}
		(*vals)[n].Disambiguation = cName + " (" + ref + ")"
	}
	return nil
}

func filterDupes(vals *[]AutocompleteValue) {
	filtered := make([]AutocompleteValue, 0, len(*vals))
Loop:
	for i := 0; i < len(*vals); i++ {
		for j := 0; j < len(*vals); j++ {
			if i == j {
				continue
			}
			if (*vals)[i] == (*vals)[j] {
				continue Loop
			}
		}
		filtered = append(filtered, (*vals)[i])
	}
	*vals = filtered
}

func (c *Calls) AutocompleteAddress(req AutocompleteAddressRequest, vals *[]AutocompleteValue) error {
	*vals = make([]AutocompleteValue, 0, MAXRETURN)
	var first, second string
	if req.Priority == 0 {
		first = "From"
		second = "To"
	} else {
		first = "To"
		second = "From"
	}
	err := c.autocomplete(vals, first, "Event", req.Partial+"%", true)
	if err != nil || len(*vals) >= MAXRETURN {
		return err
	}
	filterDupes(vals)
	notIDsOne := make([]int64, 0, MAXRETURN)
	for _, v := range *vals {
		notIDsOne = append(notIDsOne, v.ID)
	}
	preLen := len(*vals)
	err = c.autocomplete(vals, second, "Event", req.Partial+"%", true)
	if err != nil || len(*vals) >= MAXRETURN {
		return err
	}
	filterDupes(vals)
	notIDsTwo := make([]int64, 0, MAXRETURN)
	for _, v := range (*vals)[preLen:] {
		notIDsTwo = append(notIDsTwo, v.ID)
	}
	err = c.autocomplete(vals, first, "Event", "%"+req.Partial+"%", true, notIDsOne...)
	if err != nil || len(*vals) >= MAXRETURN {
		return err
	}
	filterDupes(vals)
	err = c.autocomplete(vals, second, "Event", "%"+req.Partial+"%", true, notIDsTwo...)
	if err != nil || len(*vals) >= MAXRETURN {
		return err
	}
	filterDupes(vals)
	return err
}
