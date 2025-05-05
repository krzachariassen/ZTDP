package main

import (
	"log"
	"net/http"

	"github.com/krzachariassen/ZTDP/api/server"
)

func main() {
	r := server.NewRouter()
	log.Println("Starting API on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
