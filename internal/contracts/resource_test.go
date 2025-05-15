package contracts

import (
	"testing"
)

func TestResourceTypeContract_Validate(t *testing.T) {
	tests := []struct {
		name     string
		contract ResourceTypeContract
		wantErr  bool
	}{
		{
			name: "valid resource type",
			contract: ResourceTypeContract{
				Metadata: Metadata{Name: "postgres", Owner: "platform-team"},
				Spec: ResourceTypeSpec{
					Version:         "15.0",
					DefaultTier:     "standard",
					TierOptions:     []string{"standard", "high-memory", "high-cpu"},
					ConfigTemplate:  "config/templates/postgres-config.yaml",
					AvailablePlans:  []string{"dev", "prod"},
					DefaultCapacity: "10GB",
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			contract: ResourceTypeContract{
				Metadata: Metadata{Name: "", Owner: "platform-team"},
				Spec: ResourceTypeSpec{
					Version:     "15.0",
					DefaultTier: "standard",
				},
			},
			wantErr: true,
		},
		{
			name: "missing version",
			contract: ResourceTypeContract{
				Metadata: Metadata{Name: "postgres", Owner: "platform-team"},
				Spec: ResourceTypeSpec{
					Version:     "",
					DefaultTier: "standard",
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

func TestResourceContract_Validate(t *testing.T) {
	tests := []struct {
		name     string
		contract ResourceContract
		wantErr  bool
	}{
		{
			name: "valid resource",
			contract: ResourceContract{
				Metadata: Metadata{Name: "checkout-postgres", Owner: "team-x"},
				Spec: ResourceSpec{
					Type:     "postgres",
					Version:  "15.0",
					Tier:     "standard",
					Capacity: "20GB",
					Plan:     "prod",
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			contract: ResourceContract{
				Metadata: Metadata{Name: "", Owner: "team-x"},
				Spec: ResourceSpec{
					Type:    "postgres",
					Version: "15.0",
				},
			},
			wantErr: true,
		},
		{
			name: "missing type",
			contract: ResourceContract{
				Metadata: Metadata{Name: "checkout-postgres", Owner: "team-x"},
				Spec: ResourceSpec{
					Type:    "",
					Version: "15.0",
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
