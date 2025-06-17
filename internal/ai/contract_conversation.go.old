package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/krzachariassen/ZTDP/internal/contracts"
)

// ContractConversation provides generic contract-driven conversation flows
// This is AI infrastructure - it orchestrates but doesn't contain business logic
type ContractConversation struct {
	provider AIProvider
}

// ConversationStep represents a step in the conversation flow
type ConversationStep struct {
	Type       string                 `json:"type"`        // "prompt", "validation", "completion"
	Message    string                 `json:"message"`     // Message to user
	Required   []string               `json:"required"`    // Required fields for this step
	Collected  map[string]interface{} `json:"collected"`   // Data collected so far
	Contract   contracts.Contract     `json:"contract"`    // Contract being built
	IsComplete bool                   `json:"is_complete"` // Whether conversation is complete
}

// ContractRegistry maps contract kinds to their factory functions
type ContractRegistry map[string]func() contracts.Contract

// NewContractConversation creates a new contract-driven conversation engine
func NewContractConversation(provider AIProvider) *ContractConversation {
	return &ContractConversation{
		provider: provider,
	}
}

// StartContractConversation begins a conversation for creating any contract type
func (c *ContractConversation) StartContractConversation(
	ctx context.Context,
	contractKind string,
	initialQuery string,
	registry ContractRegistry,
) (*ConversationStep, error) {
	// Create contract instance from registry
	contractFactory, exists := registry[contractKind]
	if !exists {
		return nil, fmt.Errorf("unsupported contract kind: %s", contractKind)
	}

	contract := contractFactory()

	// Extract any initial data from the query using AI
	extracted, err := c.extractDataFromQuery(ctx, initialQuery, contract)
	if err != nil {
		return nil, fmt.Errorf("failed to extract data from query: %w", err)
	}

	// Determine what's missing for contract completion
	missing := c.findMissingFields(extracted, contract)

	// Generate appropriate response based on what's missing
	step := &ConversationStep{
		Type:       "prompt",
		Collected:  extracted,
		Contract:   contract,
		IsComplete: len(missing) == 0,
		Required:   missing,
	}

	if step.IsComplete {
		step.Type = "completion"
		step.Message = fmt.Sprintf("I have all the information needed to create the %s. Shall I proceed?", contractKind)
	} else {
		step.Message, err = c.generatePromptForMissingFields(ctx, contractKind, missing, extracted)
		if err != nil {
			return nil, fmt.Errorf("failed to generate prompt: %w", err)
		}
	}

	return step, nil
}

// ContinueConversation processes user response and continues the conversation
func (c *ContractConversation) ContinueConversation(
	ctx context.Context,
	step *ConversationStep,
	userResponse string,
) (*ConversationStep, error) {
	// Extract additional data from user response
	newData, err := c.extractDataFromQuery(ctx, userResponse, step.Contract)
	if err != nil {
		return nil, fmt.Errorf("failed to extract data from response: %w", err)
	}

	// Merge with existing data
	for key, value := range newData {
		step.Collected[key] = value
	}

	// Update contract with collected data
	if err := c.populateContract(step.Contract, step.Collected); err != nil {
		return nil, fmt.Errorf("failed to populate contract: %w", err)
	}

	// Check if we now have everything needed
	missing := c.findMissingFields(step.Collected, step.Contract)

	nextStep := &ConversationStep{
		Type:       "prompt",
		Collected:  step.Collected,
		Contract:   step.Contract,
		IsComplete: len(missing) == 0,
		Required:   missing,
	}

	if nextStep.IsComplete {
		nextStep.Type = "completion"
		nextStep.Message = fmt.Sprintf("Perfect! I now have all the information needed to create the %s. Shall I proceed?", step.Contract.Kind())
	} else {
		nextStep.Message, err = c.generatePromptForMissingFields(ctx, step.Contract.Kind(), missing, step.Collected)
		if err != nil {
			return nil, fmt.Errorf("failed to generate prompt: %w", err)
		}
	}

	return nextStep, nil
}

// extractDataFromQuery uses AI to extract structured data from natural language
func (c *ContractConversation) extractDataFromQuery(
	ctx context.Context,
	query string,
	contract contracts.Contract,
) (map[string]interface{}, error) {
	systemPrompt := fmt.Sprintf(`You are a data extraction assistant for %s contracts.

Your job is to extract structured data from user queries and return ONLY valid JSON.

Contract Schema for %s:
%s

Rules:
1. Extract only data that is explicitly mentioned in the query
2. Return valid JSON with extracted fields
3. Use null for missing fields, don't make assumptions
4. Field names must match the contract schema exactly
5. Return empty object {} if no relevant data found

Example input: "Create an app called my-app with description 'My awesome app'"
Example output: {"name": "my-app", "description": "My awesome app"}`,
		contract.Kind(),
		contract.Kind(),
		c.getContractSchema(contract),
	)

	userPrompt := fmt.Sprintf("Extract data from this query: %s", query)

	response, err := c.provider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

	// Parse JSON response
	var extracted map[string]interface{}
	if err := json.Unmarshal([]byte(response), &extracted); err != nil {
		return make(map[string]interface{}), nil // Return empty map on parse error
	}

	return extracted, nil
}

// findMissingFields determines what fields are missing for contract validation
func (c *ContractConversation) findMissingFields(
	collected map[string]interface{},
	contract contracts.Contract,
) []string {
	// Create a copy of the contract and populate it with collected data
	tempContract := c.copyContract(contract)
	c.populateContract(tempContract, collected)

	// Try to validate - this will tell us what's missing
	if err := tempContract.Validate(); err != nil {
		// Parse validation error to determine missing fields
		return c.parseValidationError(err)
	}

	return []string{} // All required fields present
}

// generatePromptForMissingFields creates natural language prompts for missing data
func (c *ContractConversation) generatePromptForMissingFields(
	ctx context.Context,
	contractKind string,
	missing []string,
	collected map[string]interface{},
) (string, error) {
	if len(missing) == 0 {
		return "I have all the information needed!", nil
	}

	systemPrompt := fmt.Sprintf(`You are a conversational assistant helping users create %s contracts.

The user is missing these required fields: %v
They have already provided: %v

Generate a natural, conversational prompt asking for the missing information.
Be specific about what's needed but keep it friendly and helpful.
Don't ask for everything at once - focus on the most important missing field.`,
		contractKind,
		missing,
		collected,
	)

	userPrompt := fmt.Sprintf("Generate a prompt asking for missing fields: %v", missing)

	response, err := c.provider.CallAI(ctx, systemPrompt, userPrompt)
	if err != nil {
		// Fallback to simple prompt
		return fmt.Sprintf("I need a few more details. Could you provide the %s?", missing[0]), nil
	}

	return response, nil
}

// Helper methods for contract manipulation
func (c *ContractConversation) getContractSchema(contract contracts.Contract) string {
	// Use reflection to build schema description
	contractType := reflect.TypeOf(contract)
	if contractType.Kind() == reflect.Ptr {
		contractType = contractType.Elem()
	}

	var schema strings.Builder
	schema.WriteString(fmt.Sprintf("Kind: %s\n", contract.Kind()))
	schema.WriteString("Required fields:\n")

	// For now, we'll use a simple mapping. In production, you'd want a more sophisticated schema system
	switch contract.Kind() {
	case "application":
		schema.WriteString("- name (string): Application name\n")
		schema.WriteString("- description (string, optional): Application description\n")
		schema.WriteString("- owner (string, optional): Application owner\n")
		schema.WriteString("- tags ([]string, optional): Application tags\n")
	case "service":
		schema.WriteString("- name (string): Service name\n")
		schema.WriteString("- application (string): Linked application name\n")
		schema.WriteString("- port (int, optional): Service port\n")
		schema.WriteString("- public (bool, optional): Whether service is public\n")
	case "environment":
		schema.WriteString("- name (string): Environment name\n")
	case "resource":
		schema.WriteString("- name (string): Resource name\n")
	}

	return schema.String()
}

func (c *ContractConversation) copyContract(contract contracts.Contract) contracts.Contract {
	// Create a new instance of the same contract type
	contractType := reflect.TypeOf(contract)
	if contractType.Kind() == reflect.Ptr {
		contractType = contractType.Elem()
	}
	newContract := reflect.New(contractType).Interface().(contracts.Contract)
	return newContract
}

func (c *ContractConversation) populateContract(contract contracts.Contract, data map[string]interface{}) error {
	// Use reflection to populate contract fields from data map
	contractValue := reflect.ValueOf(contract)
	if contractValue.Kind() == reflect.Ptr {
		contractValue = contractValue.Elem()
	}

	// Populate metadata
	if metadataField := contractValue.FieldByName("Metadata"); metadataField.IsValid() && metadataField.CanSet() {
		if name, exists := data["name"]; exists {
			if nameField := metadataField.FieldByName("Name"); nameField.IsValid() && nameField.CanSet() {
				if nameStr, ok := name.(string); ok {
					nameField.SetString(nameStr)
				}
			}
		}
		if owner, exists := data["owner"]; exists {
			if ownerField := metadataField.FieldByName("Owner"); ownerField.IsValid() && ownerField.CanSet() {
				if ownerStr, ok := owner.(string); ok {
					ownerField.SetString(ownerStr)
				}
			}
		}
	}

	// Populate spec fields based on contract type
	if specField := contractValue.FieldByName("Spec"); specField.IsValid() && specField.CanSet() {
		c.populateSpec(specField, data)
	}

	return nil
}

func (c *ContractConversation) populateSpec(specField reflect.Value, data map[string]interface{}) {
	specType := specField.Type()

	for i := 0; i < specType.NumField(); i++ {
		field := specType.Field(i)
		fieldValue := specField.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		// Look for data with field name or json tag
		var dataValue interface{}
		var exists bool

		// Try exact field name first
		if dataValue, exists = data[strings.ToLower(field.Name)]; !exists {
			// Try json tag name
			if jsonTag := field.Tag.Get("json"); jsonTag != "" {
				dataValue, exists = data[jsonTag]
			}
		}

		if !exists {
			continue
		}

		// Set the field value based on its type
		switch fieldValue.Kind() {
		case reflect.String:
			if str, ok := dataValue.(string); ok {
				fieldValue.SetString(str)
			}
		case reflect.Int, reflect.Int32, reflect.Int64:
			if num, ok := dataValue.(float64); ok {
				fieldValue.SetInt(int64(num))
			}
		case reflect.Bool:
			if b, ok := dataValue.(bool); ok {
				fieldValue.SetBool(b)
			}
		case reflect.Slice:
			if slice, ok := dataValue.([]interface{}); ok {
				c.populateSliceField(fieldValue, slice)
			}
		case reflect.Map:
			if m, ok := dataValue.(map[string]interface{}); ok {
				c.populateMapField(fieldValue, m)
			}
		}
	}
}

func (c *ContractConversation) populateSliceField(fieldValue reflect.Value, data []interface{}) {
	sliceType := fieldValue.Type().Elem()
	slice := reflect.MakeSlice(fieldValue.Type(), len(data), len(data))

	for i, item := range data {
		itemValue := slice.Index(i)
		switch sliceType.Kind() {
		case reflect.String:
			if str, ok := item.(string); ok {
				itemValue.SetString(str)
			}
		}
	}

	fieldValue.Set(slice)
}

func (c *ContractConversation) populateMapField(fieldValue reflect.Value, data map[string]interface{}) {
	mapType := fieldValue.Type()
	newMap := reflect.MakeMap(mapType)

	for key, value := range data {
		keyValue := reflect.ValueOf(key)
		valueValue := reflect.ValueOf(value)
		newMap.SetMapIndex(keyValue, valueValue)
	}

	fieldValue.Set(newMap)
}

func (c *ContractConversation) parseValidationError(err error) []string {
	// Parse validation error message to extract missing fields
	errStr := err.Error()

	if strings.Contains(errStr, "name is required") {
		return []string{"name"}
	}
	if strings.Contains(errStr, "application is required") {
		return []string{"application"}
	}
	if strings.Contains(errStr, "linked application is required") {
		return []string{"application"}
	}

	// Default fallback
	return []string{"name"}
}
