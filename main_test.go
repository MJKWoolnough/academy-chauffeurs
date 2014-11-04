package main

import (
	"io/ioutil"
	"os"
	"path"

	"code.google.com/p/go-sqlite/go1/sqlite3"
)

//Functions used in testing

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
