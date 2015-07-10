package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/cheggaaa/pb"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

var (
	executeablePath = flag.String("exe", "c:\\ac\\academy-chauffeurs.exe", "path to executeable")
	serviceName     = flag.String("svc", "Academy Chauffeurs", "service name")
	url             = flag.String("url", "http://vimagination.zapto.org/ac.exe", "update url")
)

func Err(err error) bool {
	if err != nil {
		fmt.Println(err)
		r := bufio.NewReader(os.Stdin)
		r.ReadString('\n')
		return true
	}
	return false
}

func main() {
	flag.Parse()
	fname := path.Clean(*executeablePath)
	_, err := os.Stat(fname)
	if os.IsNotExist(err) {
		Err(errors.New("executable not found"))
		return
	}
	m, err := mgr.Connect()
	if Err(err) {
		return
	}
	defer m.Disconnect()
	s, err := m.OpenService(*serviceName)
	if Err(err) {
		return
	}
	_, err = s.Control(svc.Stop)
	if Err(err) {
		return
	}
	defer s.Start()
	time.Sleep(5 * time.Second)
	resp, err := http.Get(*url)
	if Err(err) {
		return
	}
	defer resp.Body.Close()
	f, err := os.Create(fname + ".new")
	if Err(err) {
		return
	}
	fmt.Println("Downloading...")
	bar := pb.New64(resp.ContentLength).SetUnits(pb.U_BYTES)
	bar.Start()
	w := io.MultiWriter(f, bar)
	_, err = io.Copy(w, resp.Body)
	f.Close()
	if Err(err) {
		os.Remove(fname + ".new")
		return
	}
	bar.FinishPrint("...done")
	Err(os.Remove(fname))
	Err(os.Rename(fname+".new", fname))
}
