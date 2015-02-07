package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"text/template"

	"github.com/MJKWoolnough/pagination"
	"github.com/MJKWoolnough/store"
)

type Server struct {
	db         *store.Store
	pages      *template.Template
	pagination pagination.Config
}

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

	s := &Server{
		db:         db,
		pages:      template.Must(template.New("templates").Funcs(template.FuncMap{"args": func(s ...string) []string { return s }}).ParseGlob("templates/*.html")),
		pagination: pagination.New(),
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
	http.HandleFunc("/autocompleteClientName", s.autocompleteClientName)

	http.HandleFunc("/companies", s.companies)
	http.HandleFunc("/addcompany", s.addCompany)
	http.HandleFunc("/updatecompany", s.updateCompany)
	http.HandleFunc("/removecompany", s.removeCompany)

	http.HandleFunc("/events", s.events)
	http.HandleFunc("/addevent", s.addEvent)
	http.HandleFunc("/updateevent", s.updateEvent)
	http.HandleFunc("/removeevent", s.removeEvent)

	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("/home/michael/Programming/Go/src/github.com/MJKWoolnough/academy-chauffeurs/resources/"))))

	setupView(s)

	http.Handle("/", http.RedirectHandler("/events", http.StatusFound))
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
