package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {
	http.HandleFunc("/", handler)

	l, err := net.Listen("unix", os.Getenv("SOCKPATH"))
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

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
