package contracts

import (
	"testing"
)

func TestApplicationContract_Validate(t *testing.T) {
	tests := []struct {
		name     string
		contract ApplicationContract
		wantErr  bool
	}{
		{
			name: "valid application",
			contract: ApplicationContract{
				Metadata: Metadata{Name: "checkout", Owner: "team-x"},
				Spec: ApplicationSpec{
					Description:  "Handles checkout flows",
					Tags:         []string{"payments", "frontend"},
					Environments: []string{"dev", "qa"},
					Lifecycle:    map[string]LifecycleDefinition{},
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			contract: ApplicationContract{
				Metadata: Metadata{Name: "", Owner: "team-x"},
				Spec: ApplicationSpec{
					Description:  "No name",
					Tags:         []string{},
					Environments: []string{"dev"},
					Lifecycle:    map[string]LifecycleDefinition{},
				},
			},
			wantErr: true,
		},
		{
			name: "missing environments",
			contract: ApplicationContract{
				Metadata: Metadata{Name: "checkout", Owner: "team-x"},
				Spec: ApplicationSpec{
					Description:  "No envs",
					Tags:         []string{},
					Environments: []string{},
					Lifecycle:    map[string]LifecycleDefinition{},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.contract.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
