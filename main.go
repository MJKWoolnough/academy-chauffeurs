package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc/jsonrpc"
	"os"
	"os/signal"

	"github.com/MJKWoolnough/httpdir"

	"golang.org/x/net/websocket"
)

func rpcHandler(conn *websocket.Conn) {
	jsonrpc.ServeConn(conn)
}

var dir http.FileSystem = httpdir.Default

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

func main() {
	const dbFName = "ac.db"
	err := backupDatabase(dbFName)
	if err != nil {
		log.Println(err)
		return
	}

	nc, err := newCalls(dbFName)

	if err != nil {
		log.Println(err)
		return
	}

	port, err := nc.getPort()
	if err != nil {
		log.Println(err)
		return
	}

	srv := new(authServeMux)

	srv.Handle("/rpc", websocket.Handler(rpcHandler))
	srv.Handle("/export", http.HandlerFunc(nc.export))
	srv.Handle("/ics", http.HandlerFunc(nc.calendar))
	srv.Handle("/", http.FileServer(dir))

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Println(err)
		return
	}

	c := make(chan os.Signal, 1)
	go func() {
		log.Println("Server Started")

		signal.Notify(c, os.Interrupt)

		<-c
		close(c)
		log.Println("Closing")
		signal.Stop(c)
		l.Close()
	}()

	err = http.Serve(l, srv)

	select {
	case <-c:
	default:
		close(c)
		log.Println(err)
	}
	nc.close()
}
