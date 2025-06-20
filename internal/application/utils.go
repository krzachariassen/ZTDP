package application

import (
	"encoding/json"
)

// mapToContract converts a map to a contract struct
func mapToContract(data map[string]interface{}, contract interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, contract)
}

// contractToMap converts a contract struct to a map
func contractToMap(contract interface{}) map[string]interface{} {
	bytes, _ := json.Marshal(contract)
	var result map[string]interface{}
	json.Unmarshal(bytes, &result)
	return result
}
