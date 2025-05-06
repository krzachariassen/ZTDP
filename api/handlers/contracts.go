package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/krzachariassen/ZTDP/internal/contracts"
	"github.com/krzachariassen/ZTDP/internal/graph"
)

var GlobalGraph *graph.GlobalGraph // Injected from main

// SubmitContract godoc
// @Summary      Submit a contract
// @Description  Submit an application or service contract to the platform
// @Tags         contracts
// @Accept       json
// @Produce      json
// @Param        contract  body      object  true  "Contract payload"
// @Success      201       {object}  map[string]interface{}
// @Failure      400       {object}  map[string]string
// @Router       /v1/contracts [post]
func SubmitContract(w http.ResponseWriter, r *http.Request) {
	var raw map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		WriteJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	kind, ok := raw["kind"].(string)
	if !ok {
		WriteJSONError(w, "Missing kind", http.StatusBadRequest)
		return
	}

	var node *graph.Node
	var err error

	switch kind {
	case "application":
		var c contracts.ApplicationContract
		if err := decodeAndValidate(raw, &c); err != nil {
			WriteJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
		node, err = graph.ResolveContract(c)

	case "service":
		var c contracts.ServiceContract
		if err := decodeAndValidate(raw, &c); err != nil {
			WriteJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
		node, err = graph.ResolveContract(c)

	default:
		WriteJSONError(w, "Unknown kind: "+kind, http.StatusBadRequest)
		return
	}

	if err != nil {
		WriteJSONError(w, err.Error(), http.StatusBadRequest)
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
