package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
)

func main() {

	//load config

	address := "127.0.0.1:8080"

	db, err := SetupDatabase("test.db")
	if err != nil {
		log.Fatalln(err)
	}

	s, err := NewServer(db)
	if err != nil {
		log.Fatalln(err)
	}

	http.HandleFunc("/drivers", s.drivers)
	http.HandleFunc("/adddriver", s.addDriver)
	http.HandleFunc("/updatedriver", s.updateDriver)
	http.HandleFunc("/removedriver", s.removeDriver)

	http.HandleFunc("/clients", s.clients)
	http.HandleFunc("/addclient", s.addClient)
	http.HandleFunc("/updateclient", s.updateClient)
	http.HandleFunc("/removeclient", s.removeClient)

	http.HandleFunc("/companies", s.companies)
	http.HandleFunc("/addcompany", s.addCompany)
	http.HandleFunc("/updatecompany", s.updateCompany)
	http.HandleFunc("/removecompany", s.removeCompany)

	l, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalln(err)
	}

	c := make(chan os.Signal, 1)

	go func() {
		defer l.Close()
		log.Println("Server Started")

		signal.Notify(c, os.Interrupt)
		defer signal.Stop(c)

		<-c
		close(c)

		log.Println("Closing")
	}()

	err := http.Serve(l, nil)
	select {
	case <-c:
	default:
		close(c)
		log.Println(err)
	}
}
