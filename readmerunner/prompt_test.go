package readmerunner

import (
	"reflect"
	"testing"
)

func TestParsePrompt(t *testing.T) {
	tc := []struct {
		name      string
		line      string
		expected  *Prompt
		expectErr bool
	}{
		{"simple", "[prompt]:# (name \"What is your name?\")", &Prompt{VarName: "name", Text: "What is your name?"}, false},
		{"complex", "[prompt]:# (name \"What is your name?\" [alice bob] alice)", &Prompt{VarName: "name", Text: "What is your name?", Options: []string{"alice", "bob"}, Default: "alice"}, false},
		{"missing variable", "[prompt]:# (\"What is your name?\")", nil, true},
		{"missing text", "[prompt]:# (name)", nil, true},
		{"missing closing bracket", "[prompt]:# (name \"What is your name?\"", nil, true},
		{"wrong order 1", "[prompt]:# (name \"What is your name?\" alice [alice, bob])", nil, true},
		{"wrong order 2", "[prompt]:# (\"What is your name?\" name)", nil, true},
		{"omit options", "[prompt]:# (name \"What is your name?\" alice)", &Prompt{VarName: "name", Text: "What is your name?", Default: "alice"}, false},
	}
	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			prompt, err := parsePrompt(tt.line)
			if err != nil {
				if !tt.expectErr {
					t.Fatalf("Unexpected error: %v", err)
				}
			} else {
				if tt.expectErr {
					t.Fatalf("Expected error, got nil")
				}

				if prompt.VarName != tt.expected.VarName {
					t.Errorf("Expected %q, got %q", tt.expected.VarName, prompt.VarName)
				}

				if prompt.Text != tt.expected.Text {
					t.Errorf("Expected %q, got %q", tt.expected.Text, prompt.Text)
				}

				if !reflect.DeepEqual(prompt.Options, tt.expected.Options) {
					t.Errorf("Expected %v, got %v", tt.expected.Options, prompt.Options)
				}

				if prompt.Default != tt.expected.Default {
					t.Errorf("Expected %q, got %q", tt.expected.Default, prompt.Default)
				}
			}
		})
	}
}

func TestProcessPrompt(t *testing.T) {
	tc := []struct {
		name      string
		prompt    []string
		responses []string
		expected  map[string]string
		expectErr bool
	}{
		{"simple", []string{"[prompt]:# (name \"What is your name?\")"}, []string{"Alice"}, map[string]string{"name": "Alice"}, false},
		{"options", []string{"[prompt]:# (name \"What is your name?\" [Alice Bob] Alice)"}, []string{"Bob"}, map[string]string{"name": "Bob"}, false},
		{"default", []string{"[prompt]:# (name \"What is your name?\" [Alice Bob] Alice)"}, []string{""}, map[string]string{"name": "Alice"}, false},
		{"invalid response", []string{"[prompt]:# (name \"What is your name?\" [Alice Bob] Alice)"}, []string{"Charlie"}, nil, true},
		{"missing response", []string{"[prompt]:# (name \"What is your name?\" [Alice Bob] Alice)"}, []string{""}, map[string]string{"name": "Alice"}, false},
		{"missing default", []string{"[prompt]:# (name \"What is your name?\" [Alice Bob])"}, []string{""}, nil, true},
		{"missing options", []string{"[prompt]:# (name \"What is your name?\")"}, []string{"Alice"}, map[string]string{"name": "Alice"}, false},
	}
	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			responses := fakePrompt(tt.responses)
			res, err := processPrompt(responses, tt.prompt)
			if err != nil {
				if !tt.expectErr {
					t.Fatalf("Unexpected error: %v", err)
				}
			} else {
				if tt.expectErr {
					t.Fatalf("Expected error, got nil")
				}

				if !reflect.DeepEqual(res, tt.expected) {
					t.Errorf("Expected %v, got %v", tt.expected, res)
				}
			}
		})
	}
}
