package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc/jsonrpc"
	"os"
	"os/signal"

	"golang.org/x/net/websocket"

	"github.com/MJKWoolnough/httpdir"
)

func rpcHandler(conn *websocket.Conn) {
	jsonrpc.ServeConn(conn)
}

var dir http.FileSystem = httpdir.Default

func main() {

	const port = 8080

	//http.Handle("/export", nil)
	//http.Handle("/ics", nil)
	http.Handle("/rpc", websocket.Handler(rpcHandler))
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
}
