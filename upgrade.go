package main

import (
	"crypto/sha1"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
)

func upgradeQueries(db *sql.DB, queries ...string) error {
	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

type eventJSON struct {
	ID                                                   int
	Note, ClientRef, InvoiceNote, InvoiceFrom, InvoiceTo string
}

type driverJSON struct {
	ID   int
	Note string
	Pos  int
	Show bool
}

func upgradeDB(db *sql.DB) error {
	var version int
	db.QueryRow("SELECT [Version] FROM [Settings];").Scan(&version)
	if version == 0 {
		log.Println("Upgrading to database version 1")
		if err := upgradeQueries(db,
			"ALTER TABLE [Settings] ADD [Version] INTEGER;",
			"ALTER TABLE [Settings] ADD [InvoiceHeader] TEXT NOT NULL DEFAULT '';",
			"ALTER TABLE [Settings] ADD [EmailSMTP] TEXT NOT NULL DEFAULT '';",
			"ALTER TABLE [Settings] ADD [EmailUsername] TEXT NOT NULL DEFAULT '';",
			"ALTER TABLE [Settings] ADD [EmailPassword] TEXT NOT NULL DEFAULT '';",
			"ALTER TABLE [Settings] ADD [EmailTemplate] TEXT NOT NULL DEFAULT '';",
			"ALTER TABLE [Event] ADD [InvoiceNote] TEXT NOT NULL DEFAULT '';",
			"ALTER TABLE [Event] ADD [InvoiceFrom] TEXT NOT NULL DEFAULT '';",
			"ALTER TABLE [Event] ADD [InvoiceTo] TEXT NOT NULL DEFAULT '';",
			"ALTER TABLE [Event] ADD [Other] TEXT NOT NULL DEFAULT '';",
			"ALTER TABLE [Driver] ADD [Pos] INTEGER NOT NULL DEFAULT 0;",
			"ALTER TABLE [Driver] ADD [Show] BOOLEAN NOT NULL DEFAULT TRUE;",
			"ALTER TABLE [Client] ADD [Email] TEXT NOT NULL DEFAULT '';",
			"CREATE TABLE [Users]([Username] TEXT, [Password] BLOB);",
			"INSERT INTO [Users]([Username], [Password]) VALUES (\"admin\", x'"+fmt.Sprintf("%x", sha1.Sum([]byte("password")))+"');",
		); err != nil {
			return err
		}
		log.Println("	Updated Table Structures")
		r, err := db.Query("SELECT [ID], [Note] FROM [Event];")
		if err != nil {
			return err
		}
		var eventTodo []eventJSON
		for r.Next() {
			var (
				id   int
				note string
			)
			if err = r.Scan(&id, &note); err != nil {
				return err
			}
			if len(note) > 0 && note[0] == '{' {
				var noteParts eventJSON
				if err = json.Unmarshal([]byte(note), &noteParts); err != nil {
					continue
				}
				noteParts.ID = id
				eventTodo = append(eventTodo, noteParts)
			}
		}

		if err = r.Close(); err != nil {
			return err
		}
		log.Println("	Updating Event table")
		eventTx, err := db.Begin()
		if err != nil {
			return err
		}
		eventUpdate, err := eventTx.Prepare("UPDATE [Event] SET [Note] = ?, [InvoiceNote] = ?, [InvoiceFrom] = ?, [InvoiceTo] = ? WHERE [ID] = ?;")
		if err != nil {
			return err
		}
		for _, noteParts := range eventTodo {
			if _, err := eventUpdate.Exec(noteParts.Note, noteParts.InvoiceNote, noteParts.InvoiceFrom, noteParts.InvoiceTo, noteParts.ID); err != nil {
				return err
			}
		}
		eventUpdate.Close()
		if err = eventTx.Commit(); err != nil {
			return err
		}

		log.Println("	Completed updating Event table")

		log.Println("	Updating Driver table")
		r, err = db.Query("SELECT [ID], [Note] FROM [Driver];")
		if err != nil {
			return err
		}
		var driverTodo []driverJSON
		for r.Next() {
			var (
				id   int
				note string
			)
			if err = r.Scan(&id, &note); err != nil {
				return err
			}
			if len(note) > 0 && note[0] == '{' {
				var noteParts driverJSON
				if err = json.Unmarshal([]byte(note), &noteParts); err != nil {
					continue
				}
				noteParts.ID = id
				driverTodo = append(driverTodo, noteParts)
			}
		}
		if err = r.Close(); err != nil {
			return err
		}
		driverTx, err := db.Begin()
		if err != nil {
			return err
		}
		driverUpdate, err := driverTx.Prepare("UPDATE [Driver] SET [Note] = ?, [Pos] = ?, [Show] = ? WHERE [ID] = ?;")
		if err != nil {
			return err
		}
		for _, noteParts := range driverTodo {
			if _, err = driverUpdate.Exec(noteParts.Note, noteParts.Pos, noteParts.Show, noteParts.ID); err != nil {
				return err
			}
		}

		driverUpdate.Close()
		if err = driverTx.Commit(); err != nil {
			return err
		}

		log.Println("	Completed updating Driver table")

		db.Exec("UPDATE [Settings] SET [Version] = 1;")
		version = 1

		log.Println("Completed updating to version 1")
	}
	if version == 1 {
		log.Println("Upgrading to database version 2")
		if err := upgradeQueries(db,
			"ALTER TABLE [Company] ADD [InvoiceNo] INTEGER;",
			"ALTER TABLE [Client] ADD [Address] TEXT NOT NULL DEFAULT '';",
		); err != nil {
			return err
		}
		db.Exec("UPDATE [Settings] SET [Version] = 2;")
		version = 2

		log.Println("Completed updating to version 2")
	}
	if version == 2 {
		log.Println("Upgrading to database version 3")
		if err := upgradeQueries(db,
			"ALTER TABLE [Event] ADD [Booker] TEXT NOT NULL DEFAULT '';",
			"ALTER TABLE [Event] ADD [FlightTime] TEXT NOT NULL DEFAULT '';",
			"ALTER TABLE [Event] ADD [ClientReference] TEXT NOT NULL DEFAULT '';",
			"UPDATE [Event] SET [ClientReference] = (SELECT [Client].[Reference] FROM [Client] WHERE [Client].[ID] = [Event].[ClientID]);",
		); err != nil {
			return err
		}

		db.Exec("UPDATE [Settings] SET [Version] = 3;")
		version = 3

		log.Println("Completed updating to version 3")
	}
	if version == 3 {
		log.Println("Upgrading to database version 4")

		if err := upgradeQueries(db,
			"CREATE TABLE [Profiles]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Name] TEXT NOT NULL DEFAULT '', [InvoiceHeader] TEXT NOT NULL DEFAULT '', [VATPercent] REAL, [AdminPercent] REAL);",
			"ALTER TABLE [Event] ADD [Profile] INTEGER DEFAULT 0;",
			"INSERT INTO [Profiles]([ID], [Name], [InvoiceHeader], [VATPercent], [AdminPercent]) SELECT 0, \"DEFAULT\", [InvoiceHeader], [VATPercent], [AdminPercent] FROM [Settings];",
			"CREATE TABLE [NEW_Settings]([TMUsername] TEXT, [TMPassword] TEXT, [TMTemplate] TEXT, [TMUseNumber] BOOLEAN DEFAULT 0 NOT NULL CHECK ([TMUseNumber] IN (0,1)), [TMFrom] TEXT, [VATPercent] REAL, [AdminPercent] REAL, [Port] INTEGER, [Unassigned] INTEGER, [AlarmTime] INTEGER, [Version] INTEGER, [InvoiceHeader] TEXT NOT NULL DEFAULT '', [EmailSMTP] TEXT NOT NULL DEFAULT '', [EmailUsername] TEXT NOT NULL DEFAULT '', [EmailPassword] TEXT NOT NULL DEFAULT '', [EmailTemplate] TEXT NOT NULL DEFAULT '');",
			"INSERT INTO [NEW_Settings] ([TMUsername], [TMPassword], [TMTemplate], [TMUseNumber], [TMFrom], [VATPercent], [AdminPercent], [Port], [Unassigned], [AlarmTime], [Version], [InvoiceHeader], [EmailSMTP], [EmailUsername], [EmailPassword], [EmailTemplate]) SELECT [TMUsername], [TMPassword], [TMTemplate], [TMUseNumber], [TMFrom], [VATPercent], [AdminPercent], [Port], [Unassigned], [AlarmTime], [Version], [InvoiceHeader], [EmailSMTP], [EmailUsername], [EmailPassword], [EmailTemplate] FROM [Settings];",
			"DROP TABLE [Settings];",
			"ALTER TABLE [NEW_Settings] RENAME TO [Settings];",
		); err != nil {
			return err
		}

		db.Exec("UPDATE [Settings] SET [Version] = 4;")
		version = 4
		log.Println("Completed updating to version 4")
	}
	return nil
}
