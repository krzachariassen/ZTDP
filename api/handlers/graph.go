package handlers

import (
	"net/http"
)

func GetGraph(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Graph details"))
}
