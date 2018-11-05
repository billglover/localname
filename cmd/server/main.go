package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("environment variable PORT is required")
	}

	http.HandleFunc("/ip", handler)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handler(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	xff := req.Header.Get("X-Forwarded-For")
	ip := net.ParseIP(xff)

	if ip == nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Fprintln(rw, ip)
}
