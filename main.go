package main

import (
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/go-sql-driver/mysql"

	"github.com/fugiman/api.clickquest.net/api"
	_ "github.com/fugiman/api.clickquest.net/handlers"
)

func main() {
	http.HandleFunc("/", api.API)

	l, err := net.Listen("tcp", "0.0.0.0:9999")
	if err != nil {
		log.Print("net.Listen:", err)
		return
	}
	defer l.Close()

	api.DB, err = sql.Open("mysql", os.Getenv("MYSQL")+"clickquest.net?parseTime=true")
	if err != nil {
		log.Print("sql.Open:", err)
		return
	}

	api.Start()

	shutdown := make(chan struct{})
	go func() {
		if err := http.Serve(l, nil); err != nil {
			log.Println("http.Serve:", err)
		}
		shutdown <- struct{}{}
	}()

	// Wait for either a signal or our server to stop
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	select {
	case <-c:
	case <-shutdown:
	}
}
