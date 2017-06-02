package main

import "database/sql"

func upgradeQueries(db *sql.DB, queries ...string) error {
	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

type eventJSON struct {
	Note, ClientRef, InvoiceNote, InvoiceFrom, InvoiceTo string
}

type driverJSON struct {
	Note string
	Pos  int
	Show bool
}

func upgradeDB(db *sql.DB) error {
	var version int
	err := db.QueryRow("SELECT [Version] FROM [Settings];").Scan(&version)
	if err != nil {
		return err
	}
	if version == 0 {
		if _, err = upgradeQueries(db,
			"ALTER TABLE [Settings] ADD [Version] INTEGER;",
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
				if err = json.Unmarshall([]byte(note), &noteParts); err != nil {
					continue
				}
				if err = db.Exec("UPDATE [Event] SET [Note] = ?, [ClientRef] = ?, [InvoiceNote] = ?, [InvoiceFrom] = ?, [InvoiceTo] = ? WHERE [ID] = ?;", noteParts.Note, noteParts.InvoiceNote, noteParts.InvoiceFrom, noteParts.InvoiceFrom, noteParts.InvoiceTo, id); err != nil {
					return err
				}
			}
		}
		if err = r.Close(); err != nil {
			return err
		}

		r, err = db.Query("SELECT [ID], [Note] FROM [Driver];")
		if err != nil {
			return err
		}
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
				if err = json.Unmarshall([]byte(note), &noteParts); err != nil {
					continue
				}
				if err = db.Exec("UPDATE [Driver] SET [Note] = ?, [Pos] = ?, [Show] = ? WHERE [ID] = ?;", noteParts.Note, noteParts.Pos, noteParts.Show, id); err != nil {
					return err
				}
			}
		}
		if err = r.Close(); err != nil {
			return err
		}

		db.Exec("UPDATE [Settings] SET [Version] = 1;")
		version = 1
	}
	return nil
}
