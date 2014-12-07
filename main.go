package main

import (
	"fmt"
	"net"
	"net/http"
)

func main() {
	//Load Config??
	db, err := SetupDatabase("test.db")
	if err != nil {
		fmt.Println(err)
		return
	}

	s, err := NewServer(db)
	if err != nil {
		fmt.Println(err)
		return
	}

	http.HandleFunc("/drivers", s.drivers)
	http.HandleFunc("/adddriver", s.addDriver)
	http.HandleFunc("/updatedriver", s.updateDriver)
	http.HandleFunc("/removedriver", s.removeDriver)

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Println(err)
		return
	}

	done := make(chan struct{})
	go func() {
		err := http.Serve(l, nil)
		if err != nil {
			fmt.Println(err)
		}
		close(done)
	}()
	//Start Browser

	OpenBrowser("http://" + l.Addr().String() + "/drivers")

	l.Close()
	<-done
}
