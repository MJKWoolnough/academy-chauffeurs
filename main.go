package main

import (
	"log"
	"net"
	"net/http"
	"net/rpc/jsonrpc"
	"os"
	"os/signal"

	_ "github.com/mxk/go-sqlite/sqlite3"
	"golang.org/x/net/websocket"
)

type file struct {
	filename, header string
}

func (f file) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", f.header)
	http.ServeFile(w, r, f.filename)
}

func rpcHandler(conn *websocket.Conn) {
	jsonrpc.ServeConn(conn)
}

func main() {
	const (
		address = "127.0.0.1:8080"
		dbFName = "test.db"
	)

	nc, err := newCalls(dbFName)

	if err != nil {
		log.Println(err)
		return
	}

	http.Handle("/", file{"page.html", "application/xhtml+xml; charset=utf-8"})
	http.Handle("/code.js", file{"code.js", "text/javascript; charset=utf-8"})
	http.Handle("/style.css", file{"style.css", "text/css; charset=utf-8"})
	http.Handle("/rpc", websocket.Handler(rpcHandler))

	l, err := net.Listen("tcp", address)
	if err != nil {
		log.Println(err)
		return
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
	nc.close()
}
