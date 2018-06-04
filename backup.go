package main

import (
	"io"
	"net/http"
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
	if _, err := os.Stat(fname); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
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

	files, err := readBackupDir(nfn)
	if err != nil {
		return err
	}
	for len(files) >= maxBackups {
		os.Remove(path.Join(backupDir, files[0]))
		files = files[1:]
	}

	return nil
}

func readBackupDir(except string) ([]string, error) {
	d, _ := os.Open(backupDir)
	fis, err := d.Readdir(-1)
	d.Close()
	if err != nil {
		return nil, err
	}
	files := make([]string, 0, maxBackups)
	regex := regexp.MustCompile("backup_[0-9]{4}-[0-9]{2}-[0-9]{2}\\.db")
	for _, file := range fis {
		filename := file.Name()
		if filename == except {
			continue
		}
		if regex.MatchString(filename) {
			files = append(files, filename)
		}
	}
	sort.Strings(files)
	return files, nil
}

func getDatabase(w http.ResponseWriter, r *http.Request) {
	files, err := readBackupDir("")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, err.Error())
		return
	}
	if len(files) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, path.Join(backupDir, files[len(files)-1]))
}
