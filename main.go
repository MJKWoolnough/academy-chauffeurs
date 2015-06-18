package main

import (
	"log"
	"net"
	"net/http"
	"net/rpc/jsonrpc"
	"os"
	"os/signal"
	"path/filepath"
	"sync"

	"golang.org/x/net/websocket"
)

type file struct {
	data   []byte
	header string
}

func (f file) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", f.header)
	w.Write(f.data)
}

var (
	lock sync.Mutex
	quit = make(chan struct{})
)

func rpcHandler(conn *websocket.Conn) {
	lock.Lock()
	close(quit)
	quit = make(chan struct{})
	myQuit := quit
	lock.Unlock()
	done := make(chan struct{})
	go func() {
		select {
		case <-myQuit:
			conn.WriteClose(4000)
		case <-done:
		}
	}()
	jsonrpc.ServeConn(conn)
	close(done)
}

func main() {
	os.Chdir(filepath.Dir(os.Args[0]))
	const (
		address = "127.0.0.1:8080"
		dbFName = "ac.db"
	)
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

	http.Handle("/", file{pageHTML, "application/xhtml+xml; charset=utf-8"})
	http.Handle("/code.js", file{codeJS, "text/javascript; charset=utf-8"})
	http.Handle("/style.css", file{styleCSS, "text/css; charset=utf-8"})
	http.Handle("/rpc", websocket.Handler(rpcHandler))
	http.Handle("/export", http.HandlerFunc(nc.export))
	http.Handle("/ics", http.HandlerFunc(nc.calendar))

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
