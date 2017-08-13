package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
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

	if err := http.Serve(l, nil); err != nil {
		log.Fatal(err)
	}
}
