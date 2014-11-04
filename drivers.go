package main

import "io"

type Driver struct {
	ID                 int
	Name, Registration string
}

func (d *DB) prepareDriverStatements() error {
	err := d.Exec("CREATE TABLE IF NOT EXISTS drivers(id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, registration TEXT);")
	if err != nil {
		return err
	}
	d.addDriver, err = d.Prepare("INSERT INTO drivers (name, registration) VALUES (?, ?);")
	if err != nil {
		return err
	}
	d.getDriver, err = d.Prepare("SELECT name, registration FROM drivers WHERE id = ?;")
	if err != nil {
		return err
	}
	d.getDrivers, err = d.Prepare("SELECT id, name, registration FROM drivers ORDER BY id ASC;")
	if err != nil {
		return err
	}
	d.editDriver, err = d.Prepare("UPDATE drivers SET name = ?, registration = ? WHERE id = ?;")
	if err != nil {
		return err
	}
	d.removeDriver, err = d.Prepare("DELETE FROM drivers WHERE id = ?")
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) SetDriver(dr Driver) (int, error) {
	d.Lock()
	defer d.Unlock()
	if dr.ID > 0 {
		return dr.ID, d.editDriver.Exec(dr.Name, dr.Registration, dr.ID)
	}
	err := d.addDriver.Exec(dr.Name, dr.Registration)
	return int(d.LastInsertId()), err
}

func (d *DB) GetDriver(id int) (Driver, error) {
	d.Lock()
	defer d.Unlock()
	var dr Driver
	err := d.getDriver.Query(id)
	if err != nil {
		dr.ID = -1
		if err == io.EOF {
			err = nil
		}
		return dr, err
	}
	err = d.getDriver.Scan(&dr.Name, &dr.Registration)
	dr.ID = id
	return dr, err
}

func (d *DB) GetDrivers() (drivers []Driver, err error) {
	d.Lock()
	defer d.Unlock()
	drivers = make([]Driver, 0, 5)
	for err = d.getDrivers.Query(); err == nil; err = d.getDrivers.Next() {
		var dTmp Driver
		d.getDrivers.Scan(&dTmp.ID, &dTmp.Name, &dTmp.Registration)
		drivers = append(drivers, dTmp)
	}
	if err == io.EOF {
		err = nil
	}
	return drivers, err
}

func (d *DB) RemoveDriver(id int) error {
	d.Lock()
	defer d.Unlock()
	return d.removeDriver.Exec(id)
}
