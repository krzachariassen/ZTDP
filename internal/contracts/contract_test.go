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
					Description: "Handles checkout flows",
					Tags:        []string{"payments", "frontend"},
					Lifecycle:   map[string]LifecycleDefinition{},
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			contract: ApplicationContract{
				Metadata: Metadata{Name: "", Owner: "team-x"},
				Spec: ApplicationSpec{
					Description: "No name",
					Tags:        []string{},
					Lifecycle:   map[string]LifecycleDefinition{},
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

func TestEnvironmentContract_Validate(t *testing.T) {
	env := EnvironmentContract{
		Metadata: Metadata{Name: "dev", Owner: "platform-team"},
		Spec:     EnvironmentSpec{Description: "Development environment"},
	}
	if err := env.Validate(); err != nil {
		t.Errorf("expected valid environment contract, got error: %v", err)
	}
}

func TestMetadataFields(t *testing.T) {
	md := Metadata{Name: "foo", Owner: "bar"}
	if md.Name != "foo" {
		t.Errorf("expected Name to be 'foo', got '%s'", md.Name)
	}
	if md.Owner != "bar" {
		t.Errorf("expected Owner to be 'bar', got '%s'", md.Owner)
	}
}

func TestLifecycleDefinitionGates(t *testing.T) {
	ld := LifecycleDefinition{Gates: []string{"test", "security"}}
	if len(ld.Gates) != 2 {
		t.Errorf("expected 2 gates, got %d", len(ld.Gates))
	}
	if ld.Gates[0] != "test" || ld.Gates[1] != "security" {
		t.Errorf("unexpected gate values: %+v", ld.Gates)
	}
}

type dummyContract struct {
	id    string
	kind  string
	md    Metadata
	valid bool
}

func (d dummyContract) ID() string            { return d.id }
func (d dummyContract) Kind() string          { return d.kind }
func (d dummyContract) GetMetadata() Metadata { return d.md }
func (d dummyContract) Validate() error {
	if !d.valid {
		return &testError{"invalid contract"}
	}
	return nil
}

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }

func TestContractInterface(t *testing.T) {
	d := dummyContract{id: "id1", kind: "kind1", md: Metadata{Name: "n", Owner: "o"}, valid: true}
	var c Contract = d
	if c.ID() != "id1" {
		t.Errorf("expected ID 'id1', got '%s'", c.ID())
	}
	if c.Kind() != "kind1" {
		t.Errorf("expected Kind 'kind1', got '%s'", c.Kind())
	}
	if c.GetMetadata().Name != "n" || c.GetMetadata().Owner != "o" {
		t.Errorf("unexpected metadata: %+v", c.GetMetadata())
	}
	if err := c.Validate(); err != nil {
		t.Errorf("expected valid contract, got error: %v", err)
	}

	d2 := dummyContract{id: "id2", kind: "kind2", md: Metadata{Name: "n2", Owner: "o2"}, valid: false}
	c = d2
	if err := c.Validate(); err == nil {
		t.Error("expected error for invalid contract, got nil")
	}
}
