package main

import (
	"io"
	"os"
	"path"
	"regexp"
	"sort"
	"time"
)

const (
	backupDir  = "backup"
	maxBackups = 5
)

func backupDatabase(fname string) error {
	nfn := "backup_" + time.Now().Format("2006-01-02") + ".db"
	if _, err := os.Stat(backupDir); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(backupDir, 0755)
		} else {
			return err
		}
	} else if _, err = os.Stat(path.Join(backupDir, nfn)); err == nil || !os.IsNotExist(err) {
		return err
	}

	nf, err := os.Create(path.Join(backupDir, nfn))
	if err != nil {
		return err
	}
	defer nf.Close()
	of, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer of.Close()
	_, err = io.Copy(nf, of)
	if err != nil {
		os.Remove(path.Join(backupDir, nfn))
		return err
	}

	d, _ := os.Open(backupDir)
	fis, err := d.Readdir(-1)
	d.Close()
	if err != nil {
		return err
	}
	files := make([]string, 0, maxBackups)
	regex := regexp.MustCompile("backup_[0-9]{4}-[0-9]{2}-[0-9]{2}\\.db")
	for _, file := range fis {
		filename := file.Name()
		if filename == nfn {
			continue
		}
		if regex.MatchString(filename) {
			files = append(files, filename)
		}
	}
	sort.Strings(files)
	for len(files) >= maxBackups {
		os.Remove(path.Join(backupDir, files[0]))
		files = files[1:]
	}

	return nil
}
