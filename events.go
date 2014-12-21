package main

import "net/http"

func (s *Server) removeEvent(w http.ResponseWriter, r *http.Request) {
	s.remove(w, r, new(Event), "event", "eventRemoveConfirmation.html", "eventRemove.html")
}
