package main

import "net/http"

type authServeMux struct {
	http.ServeMux
}

func (a *authServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if ok {
		if username != "admin" || password != "password" {
			ok = false
		}
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
