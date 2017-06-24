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

	http.Handle("/rpc", websocket.Handler(rpcHandler))
	http.Handle("/export", http.HandlerFunc(nc.export))
	http.Handle("/ics", http.HandlerFunc(nc.calendar))
	http.Handle("/", http.FileServer(dir))

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
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
