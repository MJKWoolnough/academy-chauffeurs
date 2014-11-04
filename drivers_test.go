package main

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"code.google.com/p/go-sqlite/go1/sqlite3"
)

type testDB struct {
	DB
	tempDir string
}

func NewTestDB(dirname string, initFn func(*DB) error) (*testDB, error) {
	tDir, err := ioutil.TempDir(os.TempDir(), "addTests")
	if err != nil {
		return nil, err
	}
	t := &testDB{
		tempDir: tDir,
	}
	t.Conn, err = sqlite3.Open(path.Join(t.tempDir, "test.db"))
	if err != nil {
		return nil, err
	}
	err = initFn(&t.DB)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (t *testDB) CleanUp() {
	t.Close()
	os.RemoveAll(t.tempDir)
}

func TestDriverSet(t *testing.T) {
	d, err := NewTestDB("setTests", (*DB).prepareDriverStatements)
	defer d.CleanUp()
	if err != nil {
		t.Errorf("unexpected error creating test database: %s", err)
		return
	}
	tests := []struct {
		name, registration string
	}{
		{"testDriver1", "LN09TYR"},
		{"testDriver2", "JK54BKF"},
	}

	for n, test := range tests {
		if _, err = d.SetDriver(Driver{ID: 0, Name: test.name, Registration: test.registration}); err != nil {
			t.Errorf("test %d: unexpected error: %s", n+1, err)
		}
	}
}

func TestDriverSetGet(t *testing.T) {
	d, err := NewTestDB("addGetTests", (*DB).prepareDriverStatements)
	defer d.CleanUp()
	if err != nil {
		t.Errorf("unexpected error creating test database: %s", err)
		return
	}
	tests := []Driver{
		{1, "testDriver1", "LN09TYR"},
		{2, "testDriver2", "JK54BKF"},
		{3, "", ""},
	}

	d.SetDriver(Driver{ID: 0, Name: "testDriver1", Registration: "LN09TYR"})
	d.SetDriver(Driver{ID: 0, Name: "testDriver2", Registration: "JK54BKF"})

	for n, test := range tests {
		driver, err := d.GetDriver(test.ID)
		if err != nil {
			t.Errorf("test %d: unexpected error: %s", n+1, err)
		} else if driver.Name != test.Name || driver.Registration != test.Registration {
			t.Errorf("test %d: expecting driver %q (%d chars) with registration %q (%d chars), got %q (%d chars) with %q (%d chars)", n+1, test.Name, len(test.Name), test.Registration, len(test.Registration), driver.Name, len(driver.Name), driver.Registration, len(driver.Registration))
		}
	}
}

func TestDriverAddRemove(t *testing.T) {
	d, err := NewTestDB("addRemoveTests", (*DB).prepareDriverStatements)
	defer d.CleanUp()
	if err != nil {
		t.Errorf("unexpected error creating test database: %s", err)
		return
	}
	tests := []Driver{
		{1, "testDriver1", "LN09TYR"},
		{2, "", ""},
		{3, "testDriver3", "RT56FKT"},
	}

	d.SetDriver(Driver{ID: 0, Name: "testDriver1", Registration: "LN09TYR"})
	d.SetDriver(Driver{ID: 0, Name: "testDriver2", Registration: "JK54BKF"})
	d.SetDriver(Driver{ID: 0, Name: "testDriver3", Registration: "RT56FKT"})

	if err = d.RemoveDriver(2); err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	for n, test := range tests {
		driver, err := d.GetDriver(test.ID)
		if err != nil {
			t.Errorf("test %d: unexpected error: %s", n+1, err)
		} else if driver.Name != test.Name || driver.Registration != test.Registration {
			t.Errorf("test %d: expecting driver %q (%d chars) with registration %q (%d chars), got %q (%d chars) with %q (%d chars)", n+1, test.Name, len(test.Name), test.Registration, len(test.Registration), driver.Name, len(driver.Name), driver.Registration, len(driver.Registration))
		}
	}
}

func TestDriverAddGetAll(t *testing.T) {
	d, err := NewTestDB("addGetAllTests", (*DB).prepareDriverStatements)
	defer d.CleanUp()
	if err != nil {
		t.Errorf("unexpected error creating test database: %s", err)
		return
	}
	tests := []Driver{
		{0, "testDriver1", "LN09TYR"},
		{0, "testDriver2", "JK54BKF"},
		{0, "testDriver3", "RT56FKT"},
	}

	for n, test := range tests {
		_, err = d.SetDriver(test)
		if err != nil {
			t.Errorf("test %d: unexpected error: %s", n+1, err)
			continue
		}
		drivers, err := d.GetDrivers(0, 100)
		if err != nil {
			t.Errorf("test %d: unexpected error: %s", n+1, err)
		} else if len(drivers) != n+1 {
			t.Errorf("test %d: expecting %d drivers returned, got %d", n+1, n+1, len(drivers))
		}
	}
}

func TestDriverAddUpdate(t *testing.T) {
	d, err := NewTestDB("addUpdateTests", (*DB).prepareDriverStatements)
	defer d.CleanUp()
	if err != nil {
		t.Errorf("unexpected error creating test database: %s", err)
		return
	}
	tests := []struct {
		Driver
		count int
	}{
		{Driver{0, "testDriver1", "LN09TYR"}, 1},
		{Driver{0, "testDriver2", "JK54BKF"}, 2},
		{Driver{0, "testDriver3", "RT56FKT"}, 3},
		{Driver{1, "renamedDriver1", "LN09TYR"}, 3},
		{Driver{3, "renamedDriver3", "RT56FKT"}, 3},
	}

	for n, test := range tests {
		id, err := d.SetDriver(test.Driver)
		if err != nil {
			t.Errorf("test %d: unexpected error: %s", n+1, err)
			continue
		}
		driver, err := d.GetDriver(id)
		if err != nil {
			t.Errorf("test %d: unexpected error: %s", n+1, err)
			continue
		} else if driver.Name != test.Name || driver.Registration != test.Registration {
			t.Errorf("test %d: expecting driver %q (%d chars) with registration %q (%d chars), got %q (%d chars) with %q (%d chars)", n+1, test.Name, len(test.Name), test.Registration, len(test.Registration), driver.Name, len(driver.Name), driver.Registration, len(driver.Registration))
		}
		drivers, err := d.GetDrivers(0, 100)
		if err != nil {
			t.Errorf("test %d: unexpected error: %s", n+1, err)
			continue
		} else if len(drivers) != test.count {
			t.Errorf("test %d: expecting %d drivers returned, got %d", n+1, test.count, len(drivers))
		}
	}
}
