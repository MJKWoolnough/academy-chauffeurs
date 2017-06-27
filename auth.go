package main

import (
	"net/http"
	"sync"
)

type AuthMap struct {
	sync.RWMutex
	users map[string]string
}

var authMap AuthMap

func init() {
	authMap.users = make(map[string]string)
}

// Set sets the password for the given username. Returns true if the user
// already exits.
func (a *AuthMap) Set(username, password string) bool {
	a.Lock()
	_, ok := a.users[username]
	a.users[username] = password
	a.Unlock()
	return ok
}

func (a *AuthMap) Check(username, password string) bool {
	a.RLock()
	p, ok := a.users[username]
	a.RUnlock()
	return ok && password == p
}

func (a *AuthMap) Remove(username string) {
	a.Lock()
	delete(a.users, username)
	a.Unlock()
}

type authServeMux struct {
	http.ServeMux
}

func (a *authServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if ok {
		ok = authMap.Check(username, password)
	}
	if !ok {
		w.Header().Set("WWW-Authenticate", "Basic realm=\"Enter Credentials\"")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(unauthorised)
		return
	}
	a.ServeMux.ServeHTTP(w, r)
}

var unauthorised = []byte(`<html>
	<head>
		<title>Unauthorised</title>
	</head>
	<body>
		<h1>Not Authorised</h1>
	</body>
</html>
`)
