package readmerunner

import (
	"testing"
)

func TestParseTags(t *testing.T) {
	tc := []struct {
		name      string
		line      string
		expected  []string
		expectErr bool
	}{
		{"simple", "[tags]:# (foo bar)", []string{"foo", "bar"}, false},
		{"empty", "[tags]:# ()", nil, true},
		{"whitespace", "[tags]:# (  foo  bar  )", []string{"foo", "bar"}, false},
		{"single", "[tags]:# (foo)", []string{"foo"}, false},
		{"reserved", "[tags]:# (always foo bar)", []string{"always", "foo", "bar"}, false},
		{"invalid", "[tags]:# (foo bar", nil, true},
	}
	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			tags, err := parseTags(tt.line)
			if !tt.expectErr && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if tt.expectErr && err == nil {
				t.Fatalf("Expected error, got nil")
			}
			if len(tags) != len(tt.expected) {
				t.Fatalf("Expected %v, got %v", tt.expected, tags)
			}
			for i, tag := range tags {
				if tag != tt.expected[i] {
					t.Errorf("Expected %q, got %q", tt.expected[i], tag)
				}
			}
		})
	}
}

func TestCheckSectionTag(t *testing.T) {
	tc := []struct {
		name     string
		section  []string
		run      []string
		expected bool
	}{
		{"empty", []string{}, []string{}, true},
		{"empty run", []string{"foo"}, []string{}, true},
		{"empty section", []string{}, []string{"foo"}, false},
		{"match", []string{"foo", "bar"}, []string{"bar"}, true},
		{"no match", []string{"foo", "bar"}, []string{"baz"}, false},
		{"always for non-always section", []string{"foo"}, []string{"always"}, false},
		{"always for always section", []string{"always"}, []string{"always"}, true},
		{"always for non-always input", []string{"always"}, []string{"baz"}, true},
	}
	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			if checkSectionTag(tt.section, tt.run) != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, !tt.expected)
			}
		})
	}
}
