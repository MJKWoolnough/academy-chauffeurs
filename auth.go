package main

import (
	"crypto/sha1"
	"net/http"
	"sync"

	"golang.org/x/net/websocket"
)

type UserMap struct {
	sync.RWMutex
	users map[string]uint
}

var userMap UserMap

func (u *UserMap) Add(username string) {
	u.Lock()
	u.users[username] = u.users[username] + 1
	u.Unlock()
}

func (u *UserMap) Remove(username string) {
	u.Lock()
	count := u.users[username] - 1
	if count == 0 {
		delete(u.users, username)
	} else {
		u.users[username] = count
	}
	u.Unlock()
}

func (u *UserMap) Copy() map[string]uint {
	u.RLock()
	m := make(map[string]uint, len(u.users))
	for u, c := range u.users {
		m[u] = c
	}
	u.RUnlock()
	return m
}

type userConn struct {
	websocket.Handler
}

func (u *userConn) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, _, _ := r.BasicAuth()
	userMap.Add(user)
	u.Handler.ServeHTTP(w, r)
	userMap.Remove(user)
}

type AuthMap struct {
	sync.RWMutex
	users map[string][sha1.Size]byte
}

var authMap AuthMap

func init() {
	userMap.users = make(map[string]uint)
	authMap.users = make(map[string][sha1.Size]byte)
}

// Set sets the password for the given username. Returns true if the user
// already exits.
func (a *AuthMap) Set(username string, password [sha1.Size]byte) bool {
	a.Lock()
	_, ok := a.users[username]
	a.users[username] = password
	a.Unlock()
	return ok
}

func (a *AuthMap) Check(username, password string) bool {
	pwd := sha1.Sum([]byte(password))
	a.RLock()
	p, ok := a.users[username]
	a.RUnlock()
	return ok && pwd == p
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
