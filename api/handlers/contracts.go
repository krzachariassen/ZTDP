package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

var GlobalGraph *graph.GlobalGraph // Injected from main

func SubmitContract(w http.ResponseWriter, r *http.Request) {
	var raw map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	kind, ok := raw["kind"].(string)
	if !ok {
		http.Error(w, "Missing kind", http.StatusBadRequest)
		return
	}

	var node *graph.Node
	var err error

	switch kind {
	case "application":
		var c contracts.ApplicationContract
		if err := decodeAndValidate(raw, &c); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		node, err = graph.ResolveContract(c)

	case "service":
		var c contracts.ServiceContract
		if err := decodeAndValidate(raw, &c); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		node, err = graph.ResolveContract(c)

	default:
		http.Error(w, "Unknown kind: "+kind, http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	GlobalGraph.AddNode(node)

	w.WriteHeader(http.StatusCreated)
}

func decodeAndValidate(raw map[string]interface{}, target contracts.Contract) error {
	data, err := json.Marshal(raw)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, target); err != nil {
		return err
	}
	return target.Validate()
}
