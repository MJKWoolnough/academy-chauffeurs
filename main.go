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
		log.Println(err)
		return
	}

	s, err := NewServer(db)
	if err != nil {
		log.Println(err)
		return
	}

	http.HandleFunc("/drivers", s.drivers)
	http.HandleFunc("/adddriver", s.addDriver)
	http.HandleFunc("/updatedriver", s.updateDriver)
	http.HandleFunc("/removedriver", s.removeDriver)

	l, err := net.Listen("tcp", address)
	if err != nil {
		log.Println(err)
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		err := http.Serve(l, nil)
		select {
		case <-c:
		default:
			close(c)
			log.Println(err)
		}
	}()

	log.Println("Server Started")

	<-c
	signal.Stop(c)
	close(c)

	log.Println("Closing")
	l.Close()
}
