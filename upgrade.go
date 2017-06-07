package main

import (
	"database/sql"
	"encoding/json"
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
		if err := upgradeQueries(db,
			"ALTER TABLE [Settings] ADD [Version] INTEGER;",
			"ALTER TABLE [Settings] ADD [InvoiceHeader] TEXT;",
			"ALTER TABLE [Settings] ADD [EmailSMTP] TEXT;",
			"ALTER TABLE [Settings] ADD [EmailUsername] TEXT;",
			"ALTER TABLE [Settings] ADD [EmailPassword] TEXT;",
			"ALTER TABLE [Settings] ADD [EmailTemplate] TEXT;",
			"ALTER TABLE [Event] ADD [ClientRef] TEXT;",
			"ALTER TABLE [Event] ADD [InvoiceNote] TEXT;",
			"ALTER TABLE [Event] ADD [InvoiceFrom] TEXT;",
			"ALTER TABLE [Event] ADD [InvoiceTo] TEXT;",
			"ALTER TABLE [Driver] ADD [Pos] INTEGER;",
			"ALTER TABLE [Driver] ADD [Show] BOOLEAN;",
		); err != nil {
			return err
		}
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
		eventTx, err := db.Begin()
		if err != nil {
			return err
		}
		eventUpdate, err := eventTx.Prepare("UPDATE [Event] SET [Note] = ?, [ClientRef] = ?, [InvoiceNote] = ?, [InvoiceFrom] = ?, [InvoiceTo] = ? WHERE [ID] = ?;")
		if err != nil {
			return err
		}
		for _, noteParts := range eventTodo {
			if _, err := eventUpdate.Exec(noteParts.Note, noteParts.InvoiceNote, noteParts.InvoiceFrom, noteParts.InvoiceFrom, noteParts.InvoiceTo, noteParts.ID); err != nil {
				return err
			}
		}
		eventUpdate.Close()
		if err = eventTx.Commit(); err != nil {
			return err
		}

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

		db.Exec("UPDATE [Settings] SET [Version] = 1;")
		version = 1
	}
	return nil
}
