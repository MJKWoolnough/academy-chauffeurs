package main

import (
	"fmt"
	"net/http"
)

func (s *Server) drivers(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Drivers")
}

func (s *Server) addDriver(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) removeDriver(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) updateDriver(w http.ResponseWriter, r *http.Request) {

}
