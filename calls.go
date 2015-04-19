package main

import (
	"database/sql"
	"net/rpc"
	"strings"
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
	CreateFromAddress
	CreateToAddress

	ReadDriver
	ReadCompany
	ReadClient
	ReadEvent
	ReadEventFinals
	ReadFromAddress
	ReadToAddress

	UpdateDriver
	UpdateCompany
	UpdateClient
	UpdateEvent
	UpdateEventFinals
	UpdateFromAddress
	UpdateToAddress

	DeleteDriver
	DeleteCompany
	DeleteClient
	DeleteEvent
	DeleteFromAddress
	DeleteToAddress

	DeleteDriverEvents
	DeleteClientEvents
	DeleteCompanyClients
	DeleteCompanyEvents

	GetDriverNote
	GetCompanyNote
	GetClientNote
	GetEventNote

	SetDriverNote
	SetCompanyNote
	SetClientNote
	SetEventNote

	GetSettings
	SetSettings

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
	NumEventsForDriver

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
		"[Driver]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Name] TEXT, [RegistrationNumber] TEXT, [PhoneNumber] TEXT, [Note] TEXT NOT NULL DEFAULT '', [Deleted] BOOLEAN DEFAULT 0 NOT NULL CHECK ([Deleted] IN (0,1)));",
		"[Company]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Name] TEXT, [Address] TEXT, [Note] TEXT NOT NULL DEFAULT '', [Deleted] BOOLEAN DEFAULT 0 NOT NULL CHECK ([Deleted] IN (0,1)));",
		"[Client]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [CompanyID] INTEGER, [Name] TEXT, [PhoneNumber] TEXT, [Reference] TEXT, [Note] TEXT NOT NULL DEFAULT '', [Deleted] BOOLEAN DEFAULT 0 NOT NULL CHECK ([Deleted] IN (0,1)));",
		"[Event]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [DriverID] INTEGER, [ClientID] INTEGER, [Start] INTEGER, [End] INTEGER, [From] INTEGER, [To] INTEGER, [InCar] INTEGER DEFAULT 0, [Parking] INTEGER DEFAULT 0, [Waiting] INTEGER DEFAULT 0, [Drop] INTEGER DEFAULT 0, [Miles] INTEGER DEFAULT 0, [Trip] INTEGER DEFAULT 0, [Price] INTEGER DEFAULT 0, [Sub] INTEGER DEFAULT 0, [MessageSent] BOOLEAN DEFAULT 0 NOT NULL CHECK ([MessageSent] IN (0,1)), [Note] TEXT NOT NULL DEFAULT '', [FinalsSet] BOOLEAN DEFAULT 0 NOT NULL, [Deleted] BOOLEAN DEFAULT 0 NOT NULL CHECK ([Deleted] IN (0,1)));",
		"[Settings]([TMUsername] TEXT, [TMPassword] TEXT, [TMTemplate] TEXT, [TMUseNumber] BOOLEAN DEFAULT 0 NOT NULL CHECK ([TMUseNumber] IN (0,1)), [TMFrom] TEXT, [VATPercent] REAL, [AdminPercent] REAL);",
		"[FromAddresses]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Address] TEXT, [Count] INTEGER);",
		"[ToAddresses]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Address] TEXT, [Count] INTEGER);",
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
		"INSERT INTO [FromAddresses]([Address]) VALUES (?);",
		"INSERT INTO [ToAddresses]([Address], [Count]) VALUES (?, 1);",

		// Read

		"SELECT [Name], [RegistrationNumber], [PhoneNumber] FROM [Driver] WHERE [ID] = ? AND [Deleted] = 0;",
		"SELECT [Name], [Address] FROM [Company] WHERE [ID] = ? AND [Deleted] = 0;",
		"SELECT [CompanyID], [Name], [PhoneNumber], [Reference] FROM [Client] WHERE [ID] = ? AND [Deleted] = 0;",
		"SELECT [Event].[DriverID], [Event].[ClientID], [Event].[Start], [Event].[End], [FromAddresses].[Address], [ToAddresses].[Address] FROM [Event] LEFT JOIN [FromAddresses] ON ([FromAddresses].[ID] = [Event].[From]) LEFT JOIN [ToAddresses] ON ([ToAddresses].[ID] = [Event].[To]) WHERE [Event].[ID] = ? AND [Event].[Deleted] = 0;",
		"SELECT [FinalsSet], [InCar], [Parking], [Waiting], [Drop], [Miles], [Trip], [Price], [Sub] FROM [Event] WHERE [ID] = ? AND [Deleted] = 0;",
		"SELECT [ID] FROM [FromAddresses] WHERE [Address] = ?;",
		"SELECT [ID] FROM [ToAddresses] WHERE [Address] = ?;",

		// Update

		"UPDATE [Driver] SET [Name] = ?, [RegistrationNumber] = ?, [PhoneNumber] = ? WHERE [ID] = ?;",
		"UPDATE [Company] SET [Name] = ?, [Address] = ? WHERE [ID] = ?;",
		"UPDATE [Client] SET [CompanyID] = ?, [Name] = ?, [PhoneNumber] = ?, [Reference] = ? WHERE [ID] = ?;",
		"UPDATE [Event] SET [DriverID] = ?, [ClientID] = ?, [Start] = ?, [End] = ?, [From] = ?, [To] = ? WHERE [ID] = ?;",
		"UPDATE [Event] SET [FinalsSet] = 1, [InCar] = ?, [Parking] = ?, [Waiting] = ?, [Drop] = ?, [Miles] = ?, [Trip] = ?, [Price] = ?, [Sub] = ? WHERE [ID] = ?;",
		"UPDATE [FromAddresses] SET [Count] = [Count] + 1 WHERE [ID] = ?;",
		"UPDATE [ToAddresses] SET [Count] = [Count] + 1 WHERE [ID] = ?;",

		// Delete (set deleted)

		"UPDATE [Driver] SET [Deleted] = 1 WHERE [ID] = ?;",
		"UPDATE [Company] SET [Deleted] = 1 WHERE [ID] = ?;",
		"UPDATE [Client] SET [Deleted] = 1 WHERE [ID] = ?;",
		"UPDATE [Event] SET [Deleted] = 1 WHERE [ID] = ?;",

		"UPDATE [Event] SET [Deleted] = 1 WHERE [DriverID] = ?;",
		"UPDATE [Event] SET [Deleted] = 1 WHERE [ClientID] = ?;",
		"UPDATE [Client] SET [Deleted] = 1 WHERE [CompanyID] = ?;",
		"UPDATE [Event] SET [Deleted] = 1 WHERE [ClientID] IN (SELECT [ID] FROM [Client] WHERE [CompanyID] = ?);",
		"UPDATE [FromAddresses] SET [Count] = [Count] - 1 WHERE [ID] = (SELECT [From] FROM [Event] WHERE [ID] = ?);",
		"UPDATE [ToAddresses] SET [Count] = [Count] - 1 WHERE [ID] = (SELECT [To] FROM [Event] WHERE [ID] = ?);",

		// Get Notes

		"SELECT [Note] FROM [Driver] WHERE [ID] = ?;",
		"SELECT [Note] FROM [Company] WHERE [ID] = ?;",
		"SELECT [Note] FROM [Client] WHERE [ID] = ?;",
		"SELECT [Note] FROM [Event] WHERE [ID] = ?;",

		// Set Notes

		"UPDATE [Driver] SET [Note] = ? WHERE [ID] = ?;",
		"UPDATE [Company] SET [Note] = ? WHERE [ID] = ?;",
		"UPDATE [Client] SET [Note] = ? WHERE [ID] = ?;",
		"UPDATE [Event] SET [Note] = ? WHERE [ID] = ?;",

		// Settings

		"SELECT [TMUsername], [TMPassword], [TMTemplate], [TMUseNumber], [TMFrom], [VATPercent], [AdminPercent] FROM [Settings];",
		"UPDATE [Settings] SET [TMUsername] = ?, [TMPassword] = ?, [TMTemplate] = ?, [TMUseNumber] = ?, [TMFrom] = ?, [VATPercent] = ?, [AdminPercent] = ?;",

		// Searches

		// All Drivers
		"SELECT [ID], [Name], [RegistrationNumber], [PhoneNumber] FROM [Driver] WHERE [Deleted] = 0 ORDER BY [ID] ASC;",

		// Row of Events for driver
		"SELECT [Event].[ID], [Event].[DriverID], [Event].[ClientID], [Event].[Start], [Event].[End], [FromAddresses].[Address], [ToAddresses].[Address] FROM [Event] LEFT JOIN [FromAddresses] ON ([FromAddresses].[ID] = [Event].[From]) LEFT JOIN [ToAddresses] ON ([ToAddresses].[ID] = [Event].[To]) WHERE [Event].[DriverID] = ? AND [Event].[Deleted] = 0 AND [Event].[Start] Between ? AND ? ORDER BY [Event].[Start] ASC;",

		// Row of Events for client
		"SELECT [Event].[ID], [Event].[DriverID], [Event].[ClientID], [Event].[Start], [Event].[End], [FromAddresses].[Address], [ToAddresses].[Address] FROM [Event] LEFT JOIN [FromAddresses] ON ([FromAddresses].[ID] = [Event].[From]) LEFT JOIN [ToAddresses] ON ([ToAddresses].[ID] = [Event].[To]) WHERE [Event].[ClientID] = ? AND [Event].[Deleted] = 0 AND [Event].[Start] Between ? AND ? ORDER BY [Event].[Start] ASC;",

		// Row of Events for company
		"SELECT [Event].[ID], [Event].[DriverID], [Event].[ClientID], [Event].[Start], [Event].[End], [FromAddresses].[Address], [ToAddresses].[Address] FROM [Event] LEFT JOIN [FromAddresses] ON ([FromAddresses].[ID] = [Event].[From]) LEFT JOIN [ToAddresses] ON ([ToAddresses].[ID] = [Event].[To]) WHERE [Event].[ClientID] IN (SELECT [ID] FROM [Client] WHERE [CompanyID] = ?) AND [Event].[Deleted] = 0 AND [Event].[Start] Between ? AND ? ORDER BY [Event].[Start] ASC;",

		// Event Overlaps
		"SELECT COUNT(1) FROM [Event] WHERE [ID] != ? AND [Deleted] = 0 AND [DriverID] = ? AND ([Start] <= ? AND [START] > ? OR [End] <= ?3 AND [END] > ?4);",

		// Company List
		"SELECT [ID], [Name], [Address] FROM [Company] WHERE [Deleted] = 0 ORDER BY [ID] ASC;",

		// Client List
		"SELECT [ID], [CompanyID], [Name], [PhoneNumber], [Reference] FROM [Client] WHERE [Deleted] = 0 ORDER BY [ID] ASC;",

		// Clients for company
		"SELECT [ID], [CompanyID], [Name], [PhoneNumber], [Reference] FROM [Client] WHERE [CompanyID] = ? AND [Deleted] = 0 ORDER BY [ID] ASC;",

		// Events with unsent messages
		"SELECT [Event].[ID], [Event].[DriverID], [Event].[ClientID], [Event].[Start], [Event].[End], [FromAddresses].[Address], [ToAddresses].[Address] FROM [Event] LEFT JOIN [FromAddresses] ON ([FromAddresses].[ID] = [Event].[From]) LEFT JOIN [ToAddresses] ON ([ToAddresses].[ID] = [Event].[To]) WHERE [Event].[MessageSent] = 0 AND [Event].[Start] > ? AND [Event].[Deleted] = 0 ORDER BY [Event].Start ASC;",

		// Disambiguate clients
		"SELECT [Company].[Name], [Client].[Reference] FROM [Client] JOIN [Company] ON [Client].[CompanyID] = [Company].[ID] WHERE [Client].[ID] = ?;",

		// Num Clients for company
		"SELECT COUNT(1) FROM [Client] WHERE [CompanyID] = ? AND [Deleted] = 0;",

		// Num Events for company
		"SELECT COUNT(1) FROM [Event] JOIN [Client] ON [Event].[ClientID] = [Client].[ID] WHERE [Client].[CompanyID] = ? AND [Client].[Deleted] = 0 AND [Event].[Deleted] = 0;",

		// Num Events for client
		"SELECT COUNT(1) FROM [Event] WHERE [ClientID] = ? AND [Deleted] = 0;",

		// Num Events for driver
		"SELECT COUNT(1) FROM [Event] WHERE [DriverID] = ? AND [Deleted] = 0;",
	} {
		stmt, err := db.Prepare(ps)
		if err != nil {
			return nil, err
		}
		c.statements[n] = stmt
	}

	count := 0
	err = db.QueryRow("SELECT COUNT(1) FROM [Settings]").Scan(&count)
	if err != nil {
		return nil, err
	} else if count == 0 {
		_, err = db.Exec("INSERT INTO [Settings] ([TMUsername], [TMPassword], [TMTemplate], [TMUseNumber], [TMFrom], [VATPercent], [AdminPercent]) VALUES (?, ?, ?, ?, ?, ?, ?);", "username", "password", DefaultTemplate, 1, "Academy Chauffeurs", 20, 10)
		if err != nil {
			return nil, err
		}
		setMessageVars("username", "password", DefaultTemplate, "Academy Chauffeurs", true)
	} else {
		var (
			username, password, text, from string
			useNumber                      bool
		)
		err := db.QueryRow("SELECT [TMUsername], [TMPassword], [TMTemplate], [TMUseNumber], [TMFrom] FROM [Settings];").Scan(&username, &password, &text, &useNumber, &from)
		if err != nil {
			return nil, err
		}
		setMessageVars(username, password, text, from, useNumber)
		_, err = db.Exec("DELETE FROM [FromAddresses] WHERE [Count] = 0;")
		if err != nil {
			return nil, err
		}
		_, err = db.Exec("DELETE FROM [ToAddresses] WHERE [Count] = 0;")
		if err != nil {
			return nil, err
		}
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

func (c *Calls) NumClients(id int64, num *int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.statements[NumClientsForCompany].QueryRow(id).Scan(num)
}

func (c *Calls) NumEvents(id int64, num *int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.statements[NumEventsForCompany].QueryRow(id).Scan(num)
}

func (c *Calls) NumEventsClient(id int64, num *int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.statements[NumEventsForClient].QueryRow(id).Scan(num)
}

func (c *Calls) NumEventsDriver(id int64, num *int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.statements[NumEventsForDriver].QueryRow(id).Scan(num)
}

func (c *Calls) GetDriverNote(id int64, note *string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.statements[GetDriverNote].QueryRow(id).Scan(note)
}

func (c *Calls) GetCompanyNote(id int64, note *string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.statements[GetCompanyNote].QueryRow(id).Scan(note)
}

func (c *Calls) GetClientNote(id int64, note *string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.statements[GetClientNote].QueryRow(id).Scan(note)
}

func (c *Calls) GetEventNote(id int64, note *string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.statements[GetEventNote].QueryRow(id).Scan(note)
}

type NoteID struct {
	ID   int64
	Note string
}

func (c *Calls) SetDriverNote(nid NoteID, _ *struct{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, err := c.statements[SetDriverNote].Exec(nid.Note, nid.ID)
	return err
}

func (c *Calls) SetCompanyNote(nid NoteID, _ *struct{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, err := c.statements[SetCompanyNote].Exec(nid.Note, nid.ID)
	return err
}

func (c *Calls) SetClientNote(nid NoteID, _ *struct{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, err := c.statements[SetClientNote].Exec(nid.Note, nid.ID)
	return err
}

func (c *Calls) SetEventNote(nid NoteID, _ *struct{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, err := c.statements[SetEventNote].Exec(nid.Note, nid.ID)
	return err
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
	return c.getList(UnsentMessages, is{time.Now().Unix() * 1000}, func() is {
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
	filtered := make([]AutocompleteValue, 0, cap(*vals))
Loop:
	for i, val := range *vals {
		thisVal := strings.ToLower(val.Value)
		for j := 0; j < i; j++ {
			if thisVal == strings.ToLower((*vals)[j].Value) {
				continue Loop
			}
		}
		filtered = append(filtered, val)
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
	err := c.autocomplete(vals, "Address", first+"Addresses", req.Partial+"%", true)
	if err != nil || len(*vals) >= MAXRETURN {
		return err
	}
	notIDsOne := make([]int64, 0, MAXRETURN)
	for _, v := range *vals {
		notIDsOne = append(notIDsOne, v.ID)
	}
	preLen := len(*vals)
	err = c.autocomplete(vals, "Address", second+"Addresses", req.Partial+"%", true)
	filterDupes(vals)
	if err != nil || len(*vals) >= MAXRETURN {
		return err
	}
	notIDsTwo := make([]int64, 0, MAXRETURN)
	for _, v := range (*vals)[preLen:] {
		notIDsTwo = append(notIDsTwo, v.ID)
	}
	err = c.autocomplete(vals, "Address", first+"Addresses", "%"+req.Partial+"%", true, notIDsOne...)
	filterDupes(vals)
	if err != nil || len(*vals) >= MAXRETURN {
		return err
	}
	err = c.autocomplete(vals, "Address", second+"Addresses", "%"+req.Partial+"%", true, notIDsTwo...)
	filterDupes(vals)
	if err != nil || len(*vals) >= MAXRETURN {
		return err
	}
	return err
}

type EventFinals struct {
	ID, InCar, Parking, Waiting, Drop, Miles, Trip, Price, Sub int64
	FinalsSet                                                  bool
}

func (c *Calls) GetEventFinals(id int64, e *EventFinals) error {
	return c.statements[ReadEventFinals].QueryRow(id).Scan(&e.FinalsSet, &e.InCar, &e.Parking, &e.Waiting, &e.Drop, &e.Miles, &e.Trip, &e.Price, &e.Sub)
}

func (c *Calls) SetEventFinals(e EventFinals, _ *struct{}) error {
	_, err := c.statements[UpdateEventFinals].Exec(e.InCar, e.Parking, e.Waiting, e.Drop, e.Miles, e.Trip, e.Price, e.Sub, e.ID)
	return err
}

type Settings struct {
	TMUseNumber                                bool
	TMUsername, TMPassword, TMTemplate, TMFrom string
	VATPercent, AdminPercent                   float64
}

func (c *Calls) GetSettings(_ struct{}, s *Settings) error {
	return c.statements[GetSettings].QueryRow().Scan(&s.TMUsername, &s.TMPassword, &s.TMTemplate, &s.TMUseNumber, &s.TMFrom, &s.VATPercent, &s.AdminPercent)
}

func (c *Calls) SetSettings(s Settings, errStr *string) error {
	if err := setMessageVars(s.TMUsername, s.TMPassword, s.TMTemplate, s.TMFrom, s.TMUseNumber); err != nil {
		*errStr = err.Error()
		return nil
	}
	_, err := c.statements[SetSettings].Exec(s.TMUsername, s.TMPassword, s.TMTemplate, s.TMUseNumber, s.TMFrom, s.VATPercent, s.AdminPercent)
	return err
}
