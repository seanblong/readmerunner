package readmerunner

import (
	"reflect"
	"testing"
)

var markdown = `# Title
[tags]:# (always)
some content

## Section One

[tags]:# (foo bar)

` + "```bash" + `
echo "Hello, world!"
` + "```" + `

### Subsection

[tags]:# (bar)
[prompt]:# (name "What is your name?")
`

func TestParseSections(t *testing.T) {
	tc := []struct {
		name     string
		markdown string
		start    string
		tags     []string
		expected []Section
	}{
		{"empty", "", "", []string{}, nil},
		{"all", markdown, "", []string{}, []Section{
			{Type: SectionHeader, Lines: []string{"# Title", "some content", ""}, Tags: []string{"always"}},
			{Type: SectionHeader, Lines: []string{"## Section One", "", ""}, Tags: []string{"foo", "bar"}},
			{Type: SectionCode, Lines: []string{"```bash", `echo "Hello, world!"`, "```"}, Tags: []string{"foo", "bar"}},
			{Type: SectionText, Lines: []string{""}, Tags: []string{"foo", "bar"}},
			{Type: SectionHeader, Lines: []string{"### Subsection", ""}, Tags: []string{"bar"}},
			{Type: SectionPrompt, Lines: []string{"[prompt]:# (name \"What is your name?\")"}, Tags: []string{"bar"}},
		}},
		{"start", markdown, "subsection", []string{}, []Section{
			{Type: SectionHeader, Lines: []string{"# Title", "some content", ""}, Tags: []string{"always"}},
			{Type: SectionHeader, Lines: []string{"### Subsection", ""}, Tags: []string{"bar"}},
			{Type: SectionPrompt, Lines: []string{"[prompt]:# (name \"What is your name?\")"}, Tags: []string{"bar"}},
		}},
		{"tags", markdown, "", []string{"foo"}, []Section{
			{Type: SectionHeader, Lines: []string{"# Title", "some content", ""}, Tags: []string{"always"}},
			{Type: SectionHeader, Lines: []string{"## Section One", "", ""}, Tags: []string{"foo", "bar"}},
			{Type: SectionCode, Lines: []string{"```bash", `echo "Hello, world!"`, "```"}, Tags: []string{"foo", "bar"}},
			{Type: SectionText, Lines: []string{""}, Tags: []string{"foo", "bar"}},
		}},
		{"tags start", markdown, "subsection", []string{"bar"}, []Section{
			{Type: SectionHeader, Lines: []string{"# Title", "some content", ""}, Tags: []string{"always"}},
			{Type: SectionHeader, Lines: []string{"### Subsection", ""}, Tags: []string{"bar"}},
			{Type: SectionPrompt, Lines: []string{"[prompt]:# (name \"What is your name?\")"}, Tags: []string{"bar"}},
		}},
		{"nonexistent tags existing start", markdown, "subsection", []string{"baz"}, []Section{
			{Type: SectionHeader, Lines: []string{"# Title", "some content", ""}, Tags: []string{"always"}},
		}},
		{"nonexistent start", markdown, "baz", []string{""}, nil},
		{"nonexistent tags and nonexistent start", markdown, "baz", []string{"baz"}, nil},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			sections := parseSections([]byte(tt.markdown), tt.start, tt.tags)

			if len(sections) != len(tt.expected) {
				t.Fatalf("Expected %v sections, got %v", len(tt.expected), len(sections))
			}
			for i, sec := range sections {
				if sec.Type != tt.expected[i].Type {
					t.Errorf("Expected type %v, got %v", tt.expected[i].Type, sec.Type)
				}
				if !reflect.DeepEqual(sec.Lines, tt.expected[i].Lines) {
					t.Errorf("Expected lines %v, got %v", tt.expected[i].Lines, sec.Lines)
				}
				if !reflect.DeepEqual(sec.Tags, tt.expected[i].Tags) {
					t.Errorf("Expected tags %v, got %v", tt.expected[i].Tags, sec.Tags)
				}
			}
		})
	}
}
