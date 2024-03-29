package main

import (
	"crypto/sha1"
	"database/sql"
	"errors"
	"net/rpc"
	"sort"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Driver struct {
	ID                                    int64
	Name, RegistrationNumber, PhoneNumber string
	Pos                                   int
	Show                                  bool
}

type Company struct {
	ID            int64
	Name, Address string
	Colour        uint32
}

type Client struct {
	ID, CompanyID                                int64
	Name, PhoneNumber, Reference, Email, Address string
}

type Event struct {
	ID, DriverID, ClientID                               int64
	Start, End                                           int64
	Other, From, To, ClientReference, Booker, FlightTime string
	MessageSent                                          bool
	InvoiceNote, InvoiceFrom, InvoiceTo                  string
	Profile                                              uint64
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

	DeleteDriver
	DeleteCompany
	DeleteClient
	DeleteEvent

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

	SetDriverShowPos

	GetSettings
	SetSettings

	GetProfiles
	CreateProfile
	UpdateProfile
	DeleteProfile
	DefaultProfile

	DriverList
	DriverEvents
	ClientEvents
	CompanyEvents
	DriversEvents
	EventOverlap
	CompanyList
	ClientList
	ClientForCompanyList
	UnsentMessages
	MessageSent
	DisambiguateClients

	NumClientsForCompany
	NumEventsForCompany
	NumEventsForClient
	NumEventsForDriver

	CompanyColourFromClient

	AutocompleteFromAddress
	AutocompleteToAddress

	AlarmTime
	CalendarData

	UnassignedCount
	FirstUnassigned

	GetUsers
	AddUser
	UpdateUser
	RemoveUser

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
		"[Company]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Name] TEXT, [Address] TEXT, [Note] TEXT NOT NULL DEFAULT '', [Colour] INTEGER, [Deleted] BOOLEAN DEFAULT 0 NOT NULL CHECK ([Deleted] IN (0,1)));",
		"[Client]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [CompanyID] INTEGER, [Name] TEXT, [PhoneNumber] TEXT, [Reference] TEXT, [Note] TEXT NOT NULL DEFAULT '', [Deleted] BOOLEAN DEFAULT 0 NOT NULL CHECK ([Deleted] IN (0,1)));",
		"[Event]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [DriverID] INTEGER, [ClientID] INTEGER, [Start] INTEGER, [End] INTEGER, [From] INTEGER, [To] INTEGER, [InCar] INTEGER DEFAULT 0, [Parking] INTEGER DEFAULT 0, [Waiting] INTEGER DEFAULT 0, [Drop] INTEGER DEFAULT 0, [Miles] INTEGER DEFAULT 0, [Trip] INTEGER DEFAULT 0, [DriverHours] INTEGER DEFAULT 0, [Price] INTEGER DEFAULT 0, [Sub] INTEGER DEFAULT 0, [MessageSent] BOOLEAN DEFAULT 0 NOT NULL CHECK ([MessageSent] IN (0,1)), [Note] TEXT NOT NULL DEFAULT '', [FinalsSet] BOOLEAN DEFAULT 0 NOT NULL, [Created] INTEGER, [Updated] INTEGER, [Deleted] BOOLEAN DEFAULT 0 NOT NULL CHECK ([Deleted] IN (0,1)));",
		"[Settings]([TMUsername] TEXT, [TMPassword] TEXT, [TMTemplate] TEXT, [TMUseNumber] BOOLEAN DEFAULT 0 NOT NULL CHECK ([TMUseNumber] IN (0,1)), [TMFrom] TEXT, [VATPercent] REAL, [AdminPercent] REAL, [Port] INTEGER, [Unassigned] INTEGER, [AlarmTime] INTEGER);",
		"[FromAddresses]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Address] TEXT);",
		"[ToAddresses]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Address] TEXT);",
	} {
		if _, err = db.Exec("CREATE TABLE IF NOT EXISTS " + ct); err != nil {
			return nil, err
		}
	}

	if err := upgradeDB(db); err != nil {
		return nil, err
	}

	c := &Calls{
		db: db,
	}

	for n, ps := range []string{
		// Create

		"INSERT INTO [Driver]([Name], [RegistrationNumber], [PhoneNumber]) VALUES (?, ?, ?);",
		"INSERT INTO [Company]([Name], [Address], [Colour]) VALUES (?, ?, ?);",
		"INSERT INTO [Client]([CompanyID], [Name], [PhoneNumber], [Reference], [Email], [Address]) VALUES (?, ?, ?, ?, ?, ?);",
		"INSERT INTO [Event]([DriverID], [ClientID], [Start], [End], [From], [To], [Other], [Note], [ClientReference], [Booker], [FlightTime], [Profile], [Created], [Updated]) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);",
		"INSERT INTO [FromAddresses]([Address]) VALUES (?);",
		"INSERT INTO [ToAddresses]([Address]) VALUES (?);",

		// Read

		"SELECT [Name], [RegistrationNumber], [PhoneNumber], IFNULL([Pos], 0), [Show] FROM [Driver] WHERE [ID] = ? AND [Deleted] = 0;",
		"SELECT [Name], [Address], [Colour] FROM [Company] WHERE [ID] = ? AND [Deleted] = 0;",
		"SELECT [CompanyID], [Name], [PhoneNumber], [Reference], [Email], [Address] FROM [Client] WHERE [ID] = ? AND [Deleted] = 0;",
		"SELECT [Event].[DriverID], [Event].[ClientID], [Event].[Start], [Event].[End], [Event].[InvoiceNote], [Event].[InvoiceFrom], [Event].[InvoiceTo], [Event].[Other], [Event].[ClientReference], [Event].[Booker], [Event].[FlightTime], [Event].[Profile], [FromAddresses].[Address], [ToAddresses].[Address] FROM [Event] LEFT JOIN [FromAddresses] ON ([FromAddresses].[ID] = [Event].[From]) LEFT JOIN [ToAddresses] ON ([ToAddresses].[ID] = [Event].[To]) WHERE [Event].[ID] = ? AND [Event].[Deleted] = 0;",
		"SELECT [FinalsSet], [InCar], [Parking], [Waiting], [Drop], [Miles], [Trip], [DriverHours], [Price], [Sub], [InvoiceNote], [InvoiceFrom], [InvoiceTo] FROM [Event] WHERE [ID] = ? AND [Deleted] = 0;",
		"SELECT [ID] FROM [FromAddresses] WHERE [Address] = ?;",
		"SELECT [ID] FROM [ToAddresses] WHERE [Address] = ?;",

		// Update

		"UPDATE [Driver] SET [Name] = ?, [RegistrationNumber] = ?, [PhoneNumber] = ? WHERE [ID] = ?;",
		"UPDATE [Company] SET [Name] = ?, [Address] = ?, [Colour] = ? WHERE [ID] = ?;",
		"UPDATE [Client] SET [CompanyID] = ?, [Name] = ?, [PhoneNumber] = ?, [Reference] = ?, [Email] = ?, [Address] = ? WHERE [ID] = ?;",
		"UPDATE [Event] SET [DriverID] = ?, [ClientID] = ?, [Start] = ?, [End] = ?, [From] = ?, [To] = ?, [Other] = ?, [ClientReference] = ?, [Booker] = ?, [FlightTime] = ?, [Profile] = ?, [Updated] = ? WHERE [ID] = ?;",
		"UPDATE [Event] SET [FinalsSet] = 1, [InCar] = ?, [Parking] = ?, [Waiting] = ?, [Drop] = ?, [Miles] = ?, [Trip] = ?, [DriverHours] = ?, [Price] = ?, [Sub] = ?, [InvoiceNote] = ?, [InvoiceFrom] = ?, [InvoiceTo] = ? WHERE [ID] = ?;",

		// Delete (set deleted)

		"UPDATE [Driver] SET [Deleted] = 1 WHERE [ID] = ?;",
		"UPDATE [Company] SET [Deleted] = 1 WHERE [ID] = ?;",
		"UPDATE [Client] SET [Deleted] = 1 WHERE [ID] = ?;",
		"UPDATE [Event] SET [Deleted] = 1, [Updated] = ? WHERE [ID] = ?;",

		"UPDATE [Event] SET [Deleted] = 1, [Updated] = ? WHERE [DriverID] = ?;",
		"UPDATE [Event] SET [Deleted] = 1, [Updated] = ? WHERE [ClientID] = ?;",
		"UPDATE [Client] SET [Deleted] = 1 WHERE [CompanyID] = ?;",
		"UPDATE [Event] SET [Deleted] = 1, [Updated] = ? WHERE [ClientID] IN (SELECT [ID] FROM [Client] WHERE [CompanyID] = ?);",

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

		// Set Driver Show/Pos

		"UPDATE [Driver] SET [Show] = ?, [Pos] = ? WHERE [ID] = ?;",

		// Settings

		"SELECT [TMUsername], [TMPassword], [TMTemplate], [TMUseNumber], [TMFrom], [Port], [Unassigned], [AlarmTime], [EmailSMTP], [EmailUsername], [EmailPassword], [EmailTemplate], [DefaultProfile] FROM [Settings];",
		"UPDATE [Settings] SET [TMUsername] = ?, [TMPassword] = ?, [TMTemplate] = ?, [TMUseNumber] = ?, [TMFrom] = ?, [Port] = ?, [Unassigned] = ?, [AlarmTime] = ?, [EmailSMTP] = ?, [EmailUsername] = ?, [EmailPassword] = ?, [EmailTemplate] = ?, [DefaultProfile] = ?;",

		// Profiles

		"SELECT [ID], [Name], [VATPercent], [AdminPercent], [InvoiceHeader], [InvoiceFooter] FROM [Profiles] ORDER BY [ID] ASC;",
		"INSERT INTO [Profiles]([Name], [VATPercent], [AdminPercent], [InvoiceHeader], [InvoiceFooter]) VALUES (?, ?, ?, ?, ?);",
		"UPDATE [Profiles] SET [Name] = ?, [VATPercent] = ?, [AdminPercent] = ?, [InvoiceHeader] = ?, [InvoiceFooter] = ? WHERE [ID] = ?;",
		"DELETE FROM [Profiles] WHERE [ID] = ?;",
		"UPDATE [Event] SET [Profile] = 0 WHERE Profile = ?;",

		// Searches

		// All Drivers
		"SELECT [ID], [Name], [RegistrationNumber], [PhoneNumber], IFNULL([Pos], 0), [Show] FROM [Driver] WHERE [Deleted] = 0 ORDER BY [ID] ASC;",

		// Row of Events for driver
		"SELECT [Event].[ID], [Event].[DriverID], [Event].[ClientID], [Event].[Start], [Event].[End], [Event].[Other], [Event].[ClientReference], [Event].[Booker], [Event].[FlightTime], [Event].[Profile], [FromAddresses].[Address], [ToAddresses].[Address] FROM [Event] LEFT JOIN [FromAddresses] ON ([FromAddresses].[ID] = [Event].[From]) LEFT JOIN [ToAddresses] ON ([ToAddresses].[ID] = [Event].[To]) WHERE [Event].[DriverID] = ? AND [Event].[Deleted] = 0 AND [Event].[Start] >= ? AND [Event].[Start] < ? AND ? IN(-1, [Event].[Profile]) ORDER BY [Event].[Start] ASC;",

		// Row of Events for client
		"SELECT [Event].[ID], [Event].[DriverID], [Event].[ClientID], [Event].[Start], [Event].[End], [Event].[Other], [Event].[InvoiceTo], [Event].[InvoiceFrom], [Event].[InvoiceNote], [Event].[Profile], [FromAddresses].[Address], [ToAddresses].[Address] FROM [Event] LEFT JOIN [FromAddresses] ON ([FromAddresses].[ID] = [Event].[From]) LEFT JOIN [ToAddresses] ON ([ToAddresses].[ID] = [Event].[To]) WHERE [Event].[ClientID] = ? AND [Event].[Deleted] = 0 AND [Event].[Start] >= ? AND [Event].[Start] < ? AND ? IN(-1, [Event].[Profile]) ORDER BY [Event].[Start] ASC;",

		// Row of Events for company
		"SELECT [Event].[ID], [Event].[DriverID], [Event].[ClientID], [Event].[Start], [Event].[End], [Event].[Other], [Event].[ClientReference], [Event].[InvoiceTo], [Event].[InvoiceFrom], [Event].[InvoiceNote], [Event].[Profile], [FromAddresses].[Address], [ToAddresses].[Address] FROM [Event] LEFT JOIN [FromAddresses] ON ([FromAddresses].[ID] = [Event].[From]) LEFT JOIN [ToAddresses] ON ([ToAddresses].[ID] = [Event].[To]) WHERE [Event].[ClientID] IN (SELECT [ID] FROM [Client] WHERE [CompanyID] = ?) AND [Event].[Deleted] = 0 AND [Event].[Start] >= ? AND [Event].[Start] < ? AND ? IN(-1, [Event].[Profile]) ORDER BY [Event].[Start] ASC;",

		// Row of Events for drivers
		"SELECT [Event].[ID], [Event].[DriverID], [Event].[ClientID], [Event].[Start], [Event].[End], [Event].[Other], [Event].[ClientReference], [Event].[InvoiceTo], [Event].[InvoiceFrom], [Event].[InvoiceNote], [Event].[Profile], [FromAddresses].[Address], [ToAddresses].[Address] FROM [Event] LEFT JOIN [FromAddresses] ON ([FromAddresses].[ID] = [Event].[From]) LEFT JOIN [ToAddresses] ON ([ToAddresses].[ID] = [Event].[To]) WHERE [Event].[DriverID] = ? AND [Event].[Deleted] = 0 AND [Event].[Start] >= ? AND [Event].[Start] < ? AND ? IN(-1, [Event].[Profile]) ORDER BY [Event].[Start] ASC;",

		// Event Overlaps
		"SELECT COUNT(1) FROM [Event] WHERE [ID] != ? AND [Deleted] = 0 AND [DriverID] = ? AND MAX([Start], ?) < MIN([End], ?);",

		// Company List
		"SELECT [ID], [Name], [Address], [Colour] FROM [Company] WHERE [Deleted] = 0 ORDER BY [Name] ASC;",

		// Client List
		"SELECT [ID], [CompanyID], [Name], [PhoneNumber], [Reference], [Email], [Address] FROM [Client] WHERE [Deleted] = 0 ORDER BY [Name] ASC;",

		// Clients for company
		"SELECT [ID], [CompanyID], [Name], [PhoneNumber], [Reference] FROM [Client] WHERE [CompanyID] = ? AND [Deleted] = 0 ORDER BY [Name] ASC;",

		// Events with unsent messages
		"SELECT [Event].[ID], [Event].[DriverID], [Event].[ClientID], [Event].[Start], [Event].[End], [Event].[MessageSent], [FromAddresses].[Address], [ToAddresses].[Address] FROM [Event] LEFT JOIN [FromAddresses] ON ([FromAddresses].[ID] = [Event].[From]) LEFT JOIN [ToAddresses] ON ([ToAddresses].[ID] = [Event].[To]) WHERE [Event].[Start] > ? AND [Event].[DriverID] > 0 AND [Event].[Deleted] = 0 ORDER BY [Event].Start ASC;",

		// Mark message as sent
		"UPDATE [Event] SET [MessageSent] = 1 WHERE [ID] = ?;",

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

		// Company Colour from ClientID
		"SELECT [Company].[Colour] FROM [Client] LEFT JOIN [Company] ON ([Client].[CompanyID] = [Company].[ID]) WHERE [Client].[ID] = ?;",

		// Autocomplete From Address
		"SELECT [FromAddresses].[Address] FROM [Event] LEFT JOIN [FromAddresses] ON ([FromAddresses].[ID] = [Event].[From]) WHERE [Event].[ClientID] = ? GROUP BY [Event].[From] ORDER BY COUNT(1) DESC LIMIT ?;",

		// Autocomplete To Address
		"SELECT [ToAddresses].[Address] FROM [Event] LEFT JOIN [ToAddresses] ON ([ToAddresses].[ID] = [Event].[To]) WHERE [Event].[ClientID] = ? GROUP BY [Event].[From] ORDER BY COUNT(1) DESC LIMIT ?;",

		// Alarm Time Setting
		"SELECT [AlarmTime] FROM [Settings];",

		// All event data for calendar output
		"SELECT [Event].[ID], [Event].[Start], [Event].[End], [FromAddresses].[Address], [ToAddresses].[Address], [Event].[Created], [Event].[Updated], [Event].[Note], [Event].[Other], [Driver].[Name], [Client].[Name], [Company].[Name], [Client].[PhoneNumber], [Event].[FlightTime] FROM [Event] LEFT JOIN [Driver] ON ([Driver].[ID] = [Event].[DriverID]) LEFT JOIN [Client] ON ([Client].ID = [Event].[ClientID]) LEFT JOIN [Company] ON ([Company].ID = [Client].[CompanyID]) LEFT JOIN [FromAddresses] ON ([FromAddresses].[ID] = [Event].[From]) LEFT JOIN [ToAddresses] ON ([ToAddresses].[ID] = [Event].[To]) WHERE [Event].[Start] > ? AND [Event].[Deleted] = 0;",

		// Unassigned events
		"SELECT COUNT(1) FROM [Event] WHERE [DriverID] = 0 AND [Deleted] = 0;",

		"SELECT [Start] FROM [Event] WHERE [DriverID] = 0 AND [Deleted] = 0 ORDER BY [Start] ASC LIMIT 1;",

		// Users
		"SELECT [Username], [Password] FROM [Users];",
		"INSERT INTO [Users]([Username], [Password]) VALUES (?, ?);",
		"UPDATE [Users] SET [Password] = ? WHERE [Username] = ?;",
		"DELETE FROM [Users] WHERE [Username] = ?;",
	} {
		stmt, err := db.Prepare(ps)
		if err != nil {
			return nil, errors.New(err.Error() + "\n" + ps)
		}
		c.statements[n] = stmt
	}

	count := 0
	err = db.QueryRow("SELECT COUNT(1) FROM [Settings];").Scan(&count)
	if err != nil {
		return nil, err
	} else if count == 0 {
		_, err = db.Exec("INSERT INTO [Settings] ([TMUsername], [TMPassword], [TMTemplate], [TMUseNumber], [TMFrom], [VATPercent], [AdminPercent], [Port], [Unassigned], [AlarmTime]) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);", "username", "password", DefaultTemplate, 1, "Academy Chauffeurs", 20, 10, 8080, 7, -15)
		if err != nil {
			return nil, err
		}
		setMessageVars("username", "password", DefaultTemplate, "Academy Chauffeurs", true)
		setEmailVars("server", "username", "password", DefaultTemplate)
	} else {
		var (
			username, password, text, from                           string
			useNumber                                                bool
			emailServer, emailUsername, emailPassword, emailTemplete string
		)
		err := db.QueryRow("SELECT [TMUsername], [TMPassword], [TMTemplate], [TMUseNumber], [TMFrom], [EmailSMTP], [EmailUsername], [EmailPassword], [EmailTemplate] FROM [Settings];").Scan(&username, &password, &text, &useNumber, &from, &emailServer, &emailUsername, &emailPassword, &emailTemplete)
		if err != nil {
			return nil, err
		}
		setMessageVars(username, password, text, from, useNumber)
		setEmailVars(emailServer, emailUsername, emailPassword, emailTemplete)
	}
	users, err := c.statements[GetUsers].Query()
	if err != nil {
		return nil, err
	}
	for users.Next() {
		var (
			username string
			password [sha1.Size]byte
			pwd      []byte
		)
		if err = users.Scan(&username, &pwd); err != nil {
			return nil, err
		}
		copy(password[:], pwd)
		authMap.Set(username, password)
	}
	if err = users.Close(); err != nil {
		return nil, err
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

func (c *Calls) NumClientsForCompanies(ids []int64, total *int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, id := range ids {
		var num int64
		err := c.statements[NumClientsForCompany].QueryRow(id).Scan(&num)
		if err != nil {
			return err
		}
		*total += num
	}
	return nil
}

func (c *Calls) NumEvents(id int64, num *int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.statements[NumEventsForCompany].QueryRow(id).Scan(num)
}

func (c *Calls) NumEventsForCompanies(ids []int64, total *int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, id := range ids {
		var num int64
		err := c.statements[NumEventsForCompany].QueryRow(id).Scan(&num)
		if err != nil {
			return err
		}
		*total += num
	}
	return nil
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
	ID      int64 `form:"id,post"`
	Start   int64 `form:"startTime,post"`
	End     int64 `form:"endTime,post"`
	Profile int64 `form:"profile,post"`
}

func (c *Calls) DriverEvents(f EventsFilter, events *[]Event) error {
	*events = make([]Event, 0)
	return c.getList(DriverEvents, is{f.ID, f.Start, f.End, f.Profile}, func() is {
		pos := len(*events)
		*events = append(*events, Event{})
		return is{
			&(*events)[pos].ID,
			&(*events)[pos].DriverID,
			&(*events)[pos].ClientID,
			&(*events)[pos].Start,
			&(*events)[pos].End,
			&(*events)[pos].Other,
			&(*events)[pos].ClientReference,
			&(*events)[pos].Booker,
			&(*events)[pos].FlightTime,
			&(*events)[pos].Profile,
			&(*events)[pos].From,
			&(*events)[pos].To,
		}
	})
}

func (c *Calls) ClientEvents(f EventsFilter, events *[]Event) error {
	*events = make([]Event, 0)
	return c.getList(ClientEvents, is{f.ID, f.Start, f.End, f.Profile}, func() is {
		pos := len(*events)
		*events = append(*events, Event{})
		return is{
			&(*events)[pos].ID,
			&(*events)[pos].DriverID,
			&(*events)[pos].ClientID,
			&(*events)[pos].Start,
			&(*events)[pos].End,
			&(*events)[pos].Other,
			&(*events)[pos].InvoiceTo,
			&(*events)[pos].InvoiceFrom,
			&(*events)[pos].InvoiceNote,
			&(*events)[pos].Profile,
			&(*events)[pos].From,
			&(*events)[pos].To,
		}
	})
}

func (c *Calls) CompanyEvents(f EventsFilter, events *[]Event) error {
	*events = make([]Event, 0)
	return c.getList(CompanyEvents, is{f.ID, f.Start, f.End, f.Profile}, func() is {
		pos := len(*events)
		*events = append(*events, Event{})
		return is{
			&(*events)[pos].ID,
			&(*events)[pos].DriverID,
			&(*events)[pos].ClientID,
			&(*events)[pos].Start,
			&(*events)[pos].End,
			&(*events)[pos].Other,
			&(*events)[pos].ClientReference,
			&(*events)[pos].InvoiceTo,
			&(*events)[pos].InvoiceFrom,
			&(*events)[pos].InvoiceNote,
			&(*events)[pos].Profile,
			&(*events)[pos].From,
			&(*events)[pos].To,
		}
	})
}

type CEventsFilter struct {
	IDs     []int64 `form:"id,post"`
	Start   int64   `form:"startTime,post"`
	End     int64   `form:"endTime,post"`
	Profile int64   `form:"profile,post"`
}

type sortEvents []Event

func (s sortEvents) Len() int {
	return len(s)
}

func (s sortEvents) Less(i, j int) bool {
	return s[i].Start < s[j].Start
}

func (s sortEvents) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (c *Calls) DriversEvents(f CEventsFilter, events *[]Event) error {
	return c.driverCompanyEvents(DriversEvents, f, events)
}

func (c *Calls) CompaniesEvents(f CEventsFilter, events *[]Event) error {
	return c.driverCompanyEvents(CompanyEvents, f, events)
}

func (c *Calls) driverCompanyEvents(stmtID int, f CEventsFilter, events *[]Event) error {
	*events = make([]Event, 0)
	for _, id := range f.IDs {
		err := c.getList(stmtID, is{id, f.Start, f.End, f.Profile}, func() is {
			pos := len(*events)
			*events = append(*events, Event{})
			return is{
				&(*events)[pos].ID,
				&(*events)[pos].DriverID,
				&(*events)[pos].ClientID,
				&(*events)[pos].Start,
				&(*events)[pos].End,
				&(*events)[pos].Other,
				&(*events)[pos].ClientReference,
				&(*events)[pos].InvoiceTo,
				&(*events)[pos].InvoiceFrom,
				&(*events)[pos].InvoiceNote,
				&(*events)[pos].Profile,
				&(*events)[pos].From,
				&(*events)[pos].To,
			}
		})
		if err != nil {
			return err
		}
	}
	sort.Sort(sortEvents(*events))
	return nil
}

func (c *Calls) Drivers(_ struct{}, drivers *[]Driver) error {
	*drivers = make([]Driver, 0)
	return c.getList(DriverList, is{}, func() is {
		pos := len(*drivers)
		*drivers = append(*drivers, Driver{})
		return is{
			&(*drivers)[pos].ID,
			&(*drivers)[pos].Name,
			&(*drivers)[pos].RegistrationNumber,
			&(*drivers)[pos].PhoneNumber,
			&(*drivers)[pos].Pos,
			&(*drivers)[pos].Show,
		}
	})
}

func (c *Calls) Companies(_ struct{}, companies *[]Company) error {
	*companies = make([]Company, 0)
	return c.getList(CompanyList, is{}, func() is {
		pos := len(*companies)
		*companies = append(*companies, Company{})
		return is{
			&(*companies)[pos].ID,
			&(*companies)[pos].Name,
			&(*companies)[pos].Address,
			&(*companies)[pos].Colour,
		}
	})
}

func (c *Calls) Clients(_ struct{}, clients *[]Client) error {
	*clients = make([]Client, 0)
	return c.getList(ClientList, is{}, func() is {
		pos := len(*clients)
		*clients = append(*clients, Client{})
		return is{
			&(*clients)[pos].ID,
			&(*clients)[pos].CompanyID,
			&(*clients)[pos].Name,
			&(*clients)[pos].PhoneNumber,
			&(*clients)[pos].Reference,
			&(*clients)[pos].Email,
			&(*clients)[pos].Address,
		}
	})
}

func (c *Calls) ClientsForCompany(companyID int64, clients *[]Client) error {
	*clients = make([]Client, 0)
	return c.getList(ClientForCompanyList, is{companyID}, func() is {
		pos := len(*clients)
		*clients = append(*clients, Client{})
		return is{
			&(*clients)[pos].ID,
			&(*clients)[pos].CompanyID,
			&(*clients)[pos].Name,
			&(*clients)[pos].PhoneNumber,
			&(*clients)[pos].Reference,
		}
	})
}

type sortClients []Client

func (s sortClients) Len() int {
	return len(s)
}

func (s sortClients) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

func (s sortClients) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (c *Calls) ClientsForCompanies(ids []int64, clients *[]Client) error {
	*clients = make([]Client, 0)
	for _, companyID := range ids {
		err := c.getList(ClientForCompanyList, is{companyID}, func() is {
			pos := len(*clients)
			*clients = append(*clients, Client{})
			return is{
				&(*clients)[pos].ID,
				&(*clients)[pos].CompanyID,
				&(*clients)[pos].Name,
				&(*clients)[pos].PhoneNumber,
				&(*clients)[pos].Reference,
			}
		})
		if err != nil {
			return err
		}
	}
	sort.Sort(sortClients(*clients))
	return nil
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
			&(*events)[pos].MessageSent,
			&(*events)[pos].From,
			&(*events)[pos].To,
		}
	})
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

type AutocompleteAddressRequest struct {
	ClientID int64
	Priority int
	Partial  string
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
	if req.Partial == "" {
		if req.ClientID > 0 {
			var stmt int
			if req.Priority == 0 {
				stmt = AutocompleteFromAddress
			} else {
				stmt = AutocompleteToAddress
			}
			rows, err := c.statements[stmt].Query(req.ClientID, MAXRETURN)
			if err != nil {
				return err
			}
			for rows.Next() {
				var address string
				err = rows.Scan(&address)
				if err != nil {
					return err
				}
				*vals = append(*vals, AutocompleteValue{Value: address})
			}
		}
		return nil
	}
	err := c.autocomplete(vals, "Address", first+"Addresses", req.Partial+"%", true)
	if err != nil || len(*vals) >= MAXRETURN {
		return err
	}
	notIDsOne := make([]int64, 0, MAXRETURN)
	for _, v := range *vals {
		notIDsOne = append(notIDsOne, v.ID)
	}
	filterDupes(vals)
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
	ID, InCar, Parking, Waiting, Drop, Miles, Trip, DriverHours, Price, Sub int64
	InvoiceNote, InvoiceFrom, InvoiceTo                                     string
	FinalsSet                                                               bool
}

func (c *Calls) GetEventFinals(id int64, e *EventFinals) error {
	c.mu.Lock()
	err := c.statements[ReadEventFinals].QueryRow(id).Scan(&e.FinalsSet, &e.InCar, &e.Parking, &e.Waiting, &e.Drop, &e.Miles, &e.Trip, &e.DriverHours, &e.Price, &e.Sub, &e.InvoiceNote, &e.InvoiceFrom, &e.InvoiceTo)
	c.mu.Unlock()
	return err
}

func (c *Calls) SetEventFinals(e EventFinals, _ *struct{}) error {
	c.mu.Lock()
	_, err := c.statements[UpdateEventFinals].Exec(e.InCar, e.Parking, e.Waiting, e.Drop, e.Miles, e.Trip, e.DriverHours, e.Price, e.Sub, e.InvoiceNote, e.InvoiceFrom, e.InvoiceTo, e.ID)
	c.mu.Unlock()
	return err
}

type Settings struct {
	Port, Unassigned                                       uint16
	TMUseNumber                                            bool
	TMUsername, TMPassword, TMTemplate, TMFrom             string
	UploadCalendar                                         bool
	AlarmTime                                              int
	EmailSMTP, EmailUsername, EmailPassword, EmailTemplate string
	DefaultProfile                                         uint64
	Profiles
}

func (c *Calls) GetSettings(_ struct{}, s *Settings) error {
	c.mu.Lock()
	err := c.statements[GetSettings].QueryRow().Scan(&s.TMUsername, &s.TMPassword, &s.TMTemplate, &s.TMUseNumber, &s.TMFrom, &s.Port, &s.Unassigned, &s.AlarmTime, &s.EmailSMTP, &s.EmailUsername, &s.EmailPassword, &s.EmailTemplate, &s.DefaultProfile)
	c.mu.Unlock()
	if err != nil {
		return err
	}
	return c.GetProfiles(struct{}{}, &s.Profiles)
}

func (c *Calls) SetSettings(s Settings, errStr *string) error {
	if err := setMessageVars(s.TMUsername, s.TMPassword, s.TMTemplate, s.TMFrom, s.TMUseNumber); err != nil {
		*errStr = err.Error()
		return nil
	}
	if err := setEmailVars(s.EmailSMTP, s.EmailUsername, s.EmailPassword, s.EmailTemplate); err != nil {
		*errStr = err.Error()
	}
	_, err := c.statements[SetSettings].Exec(s.TMUsername, s.TMPassword, s.TMTemplate, s.TMUseNumber, s.TMFrom, s.Port, s.Unassigned, s.AlarmTime, s.EmailSMTP, s.EmailUsername, s.EmailPassword, s.EmailTemplate, s.DefaultProfile)
	return err
}

type Profiles []Profile

type Profile struct {
	ID                           int64
	Name                         string
	VATPercent, AdminPercent     float64
	InvoiceHeader, InvoiceFooter string
}

func (c *Calls) GetProfiles(_ struct{}, ps *Profiles) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	rows, err := c.statements[GetProfiles].Query()
	if err != nil {
		return err
	}
	for rows.Next() {
		var p Profile
		if err := rows.Scan(&p.ID, &p.Name, &p.VATPercent, &p.AdminPercent, &p.InvoiceHeader, &p.InvoiceFooter); err != nil {
			return err
		}
		*ps = append(*ps, p)
	}
	return rows.Close()
}

func (c *Calls) SetProfile(p Profile, newID *int64) error {
	if p.ID < 0 {
		c.mu.Lock()
		res, err := c.statements[CreateProfile].Exec(p.Name, p.VATPercent, p.AdminPercent, p.InvoiceHeader, p.InvoiceFooter)
		if err == nil {
			*newID, err = res.LastInsertId()
		}
		c.mu.Unlock()
		return err
	} else {
		c.mu.Lock()
		_, err := c.statements[UpdateProfile].Exec(p.Name, p.VATPercent, p.AdminPercent, p.InvoiceHeader, p.InvoiceFooter, p.ID)
		c.mu.Unlock()
		*newID = p.ID
		return err
	}
}

var ErrNoRemoveDefault = errors.New("cannot remove default profile")

func (c *Calls) RemoveProfile(pid uint64, _ *struct{}) error {
	if pid == 0 {
		return ErrNoRemoveDefault
	}
	c.mu.Lock()
	_, err := c.statements[DeleteProfile].Exec(pid)
	if err == nil {
		_, err = c.statements[DefaultProfile].Exec(pid)
	}
	c.mu.Unlock()
	return err
}

func (c *Calls) CompanyColour(clientID int64, colour *uint32) error {
	c.mu.Lock()
	err := c.statements[CompanyColourFromClient].QueryRow(clientID).Scan(colour)
	c.mu.Unlock()
	return err
}

func (c *Calls) FirstUnassigned(_ struct{}, t *int64) error {
	c.mu.Lock()
	err := c.statements[FirstUnassigned].QueryRow().Scan(t)
	c.mu.Unlock()
	if err != sql.ErrNoRows {
		return err
	}
	return nil
}

func (c *Calls) UnassignedCount(_ struct{}, n *uint64) error {
	c.mu.Lock()
	err := c.statements[UnassignedCount].QueryRow().Scan(n)
	c.mu.Unlock()
	return err
}

func (c *Calls) getPort() (uint16, error) {
	var port uint16
	c.mu.Lock()
	err := c.db.QueryRow("SELECT [Port] FROM [Settings];").Scan(&port)
	c.mu.Unlock()
	return port, err
}

func (c *Calls) GetUsers(_ struct{}, u *[]string) error {
	users := make([]string, 1, len(authMap.users))
	users[0] = "admin"
	for username := range authMap.users {
		if username != "admin" {
			users = append(users, username)
		}
	}
	sort.Strings(users[1:])
	*u = users
	return nil
}

type UsernamePassword struct {
	Username, Password string
}

func (c *Calls) SetUser(up UsernamePassword, _ *struct{}) error {
	var err error
	password := sha1.Sum([]byte(up.Password))
	c.mu.Lock()
	if authMap.Set(up.Username, password) {
		_, err = c.statements[UpdateUser].Exec(password[:], up.Username)
	} else {
		_, err = c.statements[AddUser].Exec(up.Username, password[:])
	}
	c.mu.Unlock()
	return err
}

func (c *Calls) RemoveUser(username string, _ *struct{}) error {
	if username == "admin" {
		return errors.New("cannot remove admin")
	}
	authMap.Remove(username)
	c.mu.Lock()
	_, err := c.statements[RemoveUser].Exec(username)
	c.mu.Unlock()
	return err
}

func (c *Calls) UsersOnline(_ struct{}, users *map[string]uint) error {
	*users = userMap.Copy()
	return nil
}

type DriverShowPos struct {
	ID   int
	Show bool
	Pos  int
}

func (c *Calls) SetDriverPosShows(dsps []DriverShowPos, _ *struct{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("UPDATE [Driver] SET [Show] = ?, [Pos] = ? WHERE [ID] = ?;")
	if err != nil {
		return err
	}
	for _, dsp := range dsps {
		_, err := stmt.Exec(dsp.Show, dsp.Pos, dsp.ID)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
