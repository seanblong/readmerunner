package readmerunner

import (
	"bytes"
	"strings"
	"testing"
)

// fakePrompt returns predetermined responses from a slice.
func fakePrompt(responses []string) func(string) string {
	index := 0
	return func(_ string) string {
		if index >= len(responses) {
			return ""
		}
		response := responses[index]
		index++
		return response
	}
}

func TestGetRunner(t *testing.T) {
	tc := []struct {
		name      string
		supported bool
	}{
		{"bash", true},
		{"python", false},
		{"go", false},
		{"", false},
	}
	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			runner := GetRunner(tt.name)
			if (runner != nil) != tt.supported {
				t.Errorf("GetRunner(%q) = %v, want supported: %v", tt.name, runner != nil, tt.supported)
			}
		})
	}
}

func TestBashRunnerRun(t *testing.T) {
	br := &BashRunner{}
	output, err := br.Run("echo hello")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expected := "hello\n"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestPrintTOC(t *testing.T) {
	mdContent := []byte(`# Title
## Section One
Some text.
### Subsection
More text.
## Section Two
`)
	var buf bytes.Buffer
	err := PrintTOC(&buf, mdContent)
	if err != nil {
		t.Fatalf("PrintTOC returned error: %v", err)
	}
	output := buf.String()
	// Check for expected lines.
	expectedLines := []string{
		"- Title",
		"  - Section One",
		"    - Subsection",
		"  - Section Two",
	}
	for _, line := range expectedLines {
		if !strings.Contains(output, line) {
			t.Errorf("Output does not contain expected line: %q", line)
		}
	}
}

func TestRunMarkdownCodeBlock(t *testing.T) {
	// Markdown with a code block that should run.
	mdContent := []byte("# Code Run Test\n```bash\necho hello world\n```")
	tc := []struct {
		name            string
		promptResponses []string
		codeExecuted    bool
	}{
		{"Inferred Skip Code Block", []string{"", ""}, false},
		{"Explicit Skip Code Block", []string{"s", ""}, false},
		{"Run Code Block", []string{"r", ""}, true},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			prompt := fakePrompt(tt.promptResponses)
			err := RunMarkdown(mdContent, "", &buf, prompt)
			if err != nil {
				t.Errorf("RunMarkdown returned error: %v", err)
			}
			output := buf.String()
			if strings.Contains(output, "Output: hello world") != tt.codeExecuted {
				if tt.codeExecuted {
					t.Errorf("Expected code execution to contain 'Output: hello world', but got %q", output)
				} else {
					t.Error("Expected code execution to not run, but got 'Output: hello world'")
				}
			}
		})
	}
}

func TestRunMarkdownMultiplePrompts(t *testing.T) {
	mdContent := []byte(`# Heading One
Paragraph one.
## Heading Two
Paragraph two.
`)
	tc := []struct {
		name            string
		promptResponses []string
		expectedOutput  string
	}{
		{"Exit After One", []string{"exit"}, "\n# Heading One\n\nParagraph one."},
		{"Exit After Two", []string{"", "exit"}, "\n# Heading One\n\nParagraph one.\n\n## Heading Two\n\nParagraph two."},
		{"Continue to End", []string{"", "", ""}, "\n# Heading One\n\nParagraph one.\n\n## Heading Two\n\nParagraph two.\n\n> README complete!\n"},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			prompt := fakePrompt(tt.promptResponses)
			err := RunMarkdown(mdContent, "", &buf, prompt)
			if err != nil {
				t.Errorf("RunMarkdown returned error: %v", err)
			}
			output := buf.String()
			if !strings.Contains(output, tt.expectedOutput) {
				t.Errorf("Expected output to contain %q, but got %q", tt.expectedOutput, output)
			}
		})
	}
}

func TestRunMarkdownWithStartAnchor(t *testing.T) {
	mdContent := []byte(`# Title
Paragraph one.
## Section One
Paragraph two.
## Subsection
Paragraph three.
## Subsection
Oh no, a match!
`)

	tc := []struct {
		name            string
		startAnchor     string
		promptResponses []string
		contain         string
		notContain      string
	}{
		{"Existing Anchor", "section-one", []string{"exit"}, "\n## Section One\n\nParagraph two.", "# Title"},
		{"Matching Anchor", "subsection", []string{"exit"}, "\n## Subsection\n\nParagraph three.", "# Section One"},
		{"Matching Anchor Plus One", "subsection", []string{"", "exit"}, "\n## Subsection\n\nParagraph three.\n\n## Subsection\n\nOh no, a match!", "# Section One"},
		{"Non-Existing Anchor", "non-existing", []string{"exit"}, "", "# Title"},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			prompt := fakePrompt(tt.promptResponses)
			err := RunMarkdown(mdContent, tt.startAnchor, &buf, prompt)
			if err != nil {
				t.Errorf("RunMarkdown returned error: %v", err)
			}
			output := buf.String()
			if !strings.Contains(output, tt.contain) {
				t.Errorf("Expected output to contain %q, but got %q", tt.contain, output)
			}
			if tt.notContain != "" && strings.Contains(output, tt.notContain) {
				t.Errorf("Expected output to not contain %q, but got %q", tt.notContain, output)
			}
		})
	}
}

func TestComplexMarkdown(t *testing.T) {
	mdContent := []byte(`# Title
- item1
- item2

More text.
No newline.

## Subheader

> Quote section
> More content
`)

	var buf bytes.Buffer
	prompt := fakePrompt([]string{"", "exit"})
	err := RunMarkdown(mdContent, "", &buf, prompt)
	if err != nil {
		t.Errorf("RunMarkdown returned error: %v", err)
	}
	got := buf.String()
	want := "\n# Title\n\n- item1\n- item2\n\nMore text.No newline.\n\n## Subheader\n\n> Quote section\n> More content\n"
	if got != want {
		t.Errorf("RunMarkdown output mismatch.\nGot:\n%s\nWant:\n%s", got, want)
	}
}
