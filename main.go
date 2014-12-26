package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/MJKWoolnough/store"
)

func main() {

	//load config

	const address = "127.0.0.1:8080"
	const dbFName = "test.db"

	db, err := store.NewStore(dbFName)
	if err != nil {
		log.Fatalln(err)
	}
	err = db.Register(new(Driver))
	if err != nil {
		log.Fatalln(err)
	}
	err = db.Register(new(Company))
	if err != nil {
		log.Fatalln(err)
	}
	err = db.Register(new(Client))
	if err != nil {
		log.Fatalln(err)
	}
	err = db.Register(new(Event))
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
	http.HandleFunc("/autocompleteClientCompanyName", s.autocompleteCompanyName)

	http.HandleFunc("/companies", s.companies)
	http.HandleFunc("/addcompany", s.addCompany)
	http.HandleFunc("/updatecompany", s.updateCompany)
	http.HandleFunc("/removecompany", s.removeCompany)

	http.HandleFunc("/events", s.events)
	http.HandleFunc("/addEvent", s.addEvent)
	http.HandleFunc("/removeevent", s.removeEvent)

	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("/home/michael/Programming/Go/src/github.com/MJKWoolnough/academy-chauffeurs/resources/"))))

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

	err = http.Serve(l, nil)
	select {
	case <-c:
	default:
		close(c)
		log.Println(err)
	}
	db.Close()
}
