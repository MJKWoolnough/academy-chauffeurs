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

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

var (
	executeablePath = flag.String("exe", "c:\\ac\\academy-chauffeurs.exe", "path to executeable")
	serviceName     = flag.String("svc", "Academy Chauffeurs", "service name")
	url             = flag.String("url", "http://vimagination.zapto.org/ac.exe", "update url")
)

func Err(err error) {
	if err != nil {
		fmt.Println(err)
		r := bufio.NewReader(os.Stdin)
		r.ReadString('\n')
		os.Exit(1)
	}
}

func main() {
	flag.Parse()
	fname := path.Clean(*executeablePath)
	_, err := os.Stat(fname)
	if os.IsNotExist(err) {
		Err(errors.New("executable not found"))
	}
	m, err := mgr.Connect()
	Err(err)
	defer m.Disconnect()
	s, err := m.OpenService(*serviceName)
	Err(err)
	_, err = s.Control(svc.Stop)
	Err(err)
	defer s.Start()
	time.Sleep(5 * time.Second)
	resp, err := http.Get(*url)
	Err(err)
	defer resp.Body.Close()
	f, err := os.Create(fname)
	Err(err)
	defer f.Close()
	fmt.Print("Downloading...")
	_, err = io.Copy(f, resp.Body)
	Err(err)
	fmt.Println(" ...done!")
}
