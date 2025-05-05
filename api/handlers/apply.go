package handlers

import (
	"net/http"
)

func ApplyGraph(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Graph applied"))
}
