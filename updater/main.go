package main

import (
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
	executeablePath = flag.String("exe", "c:\\ac\\ac.exe", "path to executeable")
	serviceName     = flag.String("svc", "Academy Chauffeurs", "service name")
	url             = flag.String("url", "http://vimagination.zapto.org/ac.exe", "update url")
)

func main() {
	flag.Parse()
	fname := path.Clean(*executeablePath)
	_, err := os.Stat(fname)
	if os.IsNotExist(err) {
		fmt.Println("executeable not found")
		return
	}
	m, err := mgr.Connect()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer m.Disconnect()
	s, err := m.OpenService(*serviceName)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = s.Control(svc.Stop)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer s.Control(svc.Start)
	time.Sleep(5 * time.Second)
	resp, err := http.Get(*url)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Close()
	f, err := os.Open(fname)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	if _, err = io.Copy(f, resp.Body); err != nil {
		fmt.Println(err)
		return
	}
}
