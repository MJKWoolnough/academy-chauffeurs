package main

import (
	"log"
	"net/http"
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

	err = http.ListenAndServe(address, nil)
	if err != nil {
		log.Println(err)
	}
}
