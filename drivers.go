package main

type Driver struct {
	ID                 int
	Name, Registration string
}

func (d *DB) prepareDriverStatements() error {
	var err error
	d.addDriver, err = d.Prepare("INSERT INTO drivers (name, registration) VALUES (?, ?)")
	if err != nil {
		return err
	}
	d.getDriver, err = d.Prepare("SELECT id, name, registration FROM drivers WHERE id = ?")
	if err != nil {
		return err
	}
	d.getDriver, err = d.Prepare("SELECT id, name, registration FROM drivers ORDER BY id ASC")
	if err != nil {
		return err
	}
	d.editDriver, err = d.Prepare("UPDATE drivers SET name = ?, registration = ? WHERE id = ?")
	if err != nil {
		return err
	}
	d.removeDriver, err = d.Prepare("DELETE FROM drivers WHERE id = ?")
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) AddDriver(name, registration string) (int, error) {
	d.Lock()
	defer d.Unlock()
	err := d.addDriver.Exec(name, registration)
	return int(d.LastInsertId()), err
}

func (d *DB) GetDriver(id int) (Driver, error) {
	d.Lock()
	defer d.Unlock()
	var dr Driver
	err := d.getDriver.Query(id)
	if err != nil {
		return dr, nil
	}
	err = d.getDriver.Next()
	if err != nil {
		return dr, nil
	}
	err = d.getDriver.Scan(dr.ID, dr.Name, dr.Registration)
	return dr, err
}

func (d *DB) GetDrivers() (drivers []Driver, err error) {
	d.Lock()
	defer d.Unlock()
	drivers = make([]Driver, 0, 5)
	for err = d.getDrivers.Query(); err == nil; err = d.getDrivers.Next() {
		var dTmp Driver
		d.getDrivers.Scan(dTmp.ID, dTmp.Name, dTmp.Registration)
		drivers = append(drivers, dTmp)
	}
	return drivers, err
}

func (d *DB) EditDriver(dr Driver) error {
	d.Lock()
	defer d.Unlock()
	return d.editDriver.Exec(dr.Name, dr.Registration, dr.ID)
}

func (d *DB) RemoveDriver(dr Driver) error {
	d.Lock()
	defer d.Unlock()
	return d.removeDriver.Exec(dr.ID)
}
