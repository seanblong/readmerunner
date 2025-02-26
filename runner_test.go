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
	runner := GetRunner("bash")
	if runner == nil {
		t.Error("Expected BashRunner for language 'bash', got nil")
	}
	runner = GetRunner("python")
	if runner != nil {
		t.Error("Expected nil for language 'python', got non-nil")
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

func TestRunMarkdownExitEarly(t *testing.T) {
	// Markdown with one heading and one paragraph.
	mdContent := []byte(`# Test Heading
This is a test paragraph.
`)
	var buf bytes.Buffer
	// Simulate user typing "exit" at the first prompt.
	prompt := fakePrompt([]string{"exit"})
	err := RunMarkdown(mdContent, "", &buf, prompt)
	if err != nil {
		t.Errorf("RunMarkdown returned error: %v", err)
	}
	// Check that the output contains the heading.
	if !strings.Contains(buf.String(), "Test Heading") {
		t.Errorf("Output does not contain heading: %q", buf.String())
	}
}

func TestRunMarkdownCodeBlockSkip(t *testing.T) {
	// Markdown with a code block that should be skipped.
	mdContent := []byte("# Code Test\nSome description.\n```bash\necho skipped\n```")

	var buf bytes.Buffer
	// Simulate user responses:
	//  - For heading: press Enter ("")
	//  - For code block: choose "s" (skip)
	//  - For following paragraph: press Enter ("")
	prompt := fakePrompt([]string{"", "s", ""})
	err := RunMarkdown(mdContent, "", &buf, prompt)
	if err != nil {
		t.Errorf("RunMarkdown returned error: %v", err)
	}
	output := buf.String()
	// Check that the code block is printed but not executed (i.e. no "Running code..." line).
	if !strings.Contains(output, "```bash") {
		t.Error("Output does not contain code block header")
	}
	if strings.Contains(output, "Running code...") {
		t.Error("Output should not contain running code message for a skipped code block")
	}
}

func TestRunMarkdownCodeBlockRun(t *testing.T) {
	// Markdown with a code block that should run.
	mdContent := []byte("# Code Run Test\n```bash\necho run\n```")

	var buf bytes.Buffer
	// Simulate user responses:
	//  - For heading: press Enter ("")
	//  - For code block: choose "r" (run)
	//  - After execution: press Enter ("")
	prompt := fakePrompt([]string{"r", ""})
	err := RunMarkdown(mdContent, "", &buf, prompt)
	if err != nil {
		t.Errorf("RunMarkdown returned error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "Output: run") {
		t.Error(output)
		t.Error("Output should contain the result of the code execution")
	}
}

func TestRunMarkdownMultiplePrompts(t *testing.T) {
	mdContent := []byte(`# Heading One
Paragraph one.
## Heading Two
Paragraph two.
`)
	var buf bytes.Buffer
	// Simulate responses:
	//  - For first heading: Enter ("")
	//  - For first paragraph: Enter ("")
	//  - For second heading: Enter ("")
	//  - For second paragraph: "exit" to stop further processing.
	prompt := fakePrompt([]string{"", "", "", "exit"})
	err := RunMarkdown(mdContent, "", &buf, prompt)
	if err != nil {
		t.Errorf("RunMarkdown returned error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "Heading One") || !strings.Contains(output, "Heading Two") {
		t.Error("Output should contain both headings")
	}
}
