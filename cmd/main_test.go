package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestRunMain_NoArgs(t *testing.T) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	// Call runMain with no arguments.
	exitCode := runMain([]string{}, strings.NewReader(""), stdout, stderr)
	if exitCode == 0 {
		t.Errorf("Expected non-zero exit code when no README is provided")
	}
	if !strings.Contains(stderr.String(), "Usage:") {
		t.Errorf("Expected usage message in stderr, got: %s", stderr.String())
	}
}

func TestRunMain_TOCMode(t *testing.T) {
	// Create a temporary README file with some headings.
	tmpFile, err := os.CreateTemp("", "README_toc_*.md")
	if err != nil {
		t.Fatalf("Error creating temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	content := "# Title\n\nSome introduction.\n\n## Section One\nMore details.\n"
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("Error writing to temp file: %v", err)
	}
	tmpFile.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	// Run in TOC mode.
	exitCode := runMain([]string{"--toc", tmpFile.Name()}, strings.NewReader(""), stdout, stderr)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 in TOC mode, got %d", exitCode)
	}
	got := stdout.String()
	if !strings.Contains(got, "Title") || !strings.Contains(got, "Section One") {
		t.Errorf("Expected TOC output to contain headings, got: %s", got)
	}
}

func TestRunMain_RunModeWithExit(t *testing.T) {
	// Create a temporary README file with some content.
	tmpFile, err := os.CreateTemp("", "README_run_*.md")
	if err != nil {
		t.Fatalf("Error creating temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	content := "# Intro\n\nWelcome to the README.\n## Section One"
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("Error writing to temp file: %v", err)
	}
	tmpFile.Close()

	// Simulate user input by providing "exit\n" so that interactive mode quits.
	stdin := strings.NewReader("exit\n")
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	exitCode := runMain([]string{tmpFile.Name()}, stdin, stdout, stderr)
	// In our runMain, a normal exit returns 0.
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 in run mode, got %d", exitCode)
	}
	got := stdout.String()
	if !strings.Contains(got, "Intro") {
		t.Errorf("Expected output to contain 'Intro', got: %s", got)
	}
	if strings.Contains(got, "README complete!") {
		t.Errorf("Expected output should not include final message 'README complete!', got: %s", got)
	}
}

func TestRunMain_InvalidLogFile(t *testing.T) {
	// Create a temporary README file.
	tmpFile, err := os.CreateTemp("", "README_invalid_log_*.md")
	if err != nil {
		t.Fatalf("Error creating temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	content := "# Intro\n\nContent.\n"
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("Error writing to temp file: %v", err)
	}
	tmpFile.Close()

	// Use a log file path that's likely invalid.
	invalidLogFile := "/invalid/path/readme-runner.log"
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	exitCode := runMain([]string{"--log", invalidLogFile, tmpFile.Name()}, strings.NewReader(""), stdout, stderr)
	if exitCode == 0 {
		t.Errorf("Expected non-zero exit code when log file cannot be opened")
	}
	if !strings.Contains(stderr.String(), "Error opening log file") {
		t.Errorf("Expected error message about log file, got: %s", stderr.String())
	}
}

func TestRunMain_StartAnchor(t *testing.T) {
	// Create a temporary README file with some content.
	tmpFile, err := os.CreateTemp("", "README_start_anchor_*.md")
	if err != nil {
		t.Fatalf("Error creating temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	content := "# Intro\n\nWelcome to the README.\n## Section One\nDetails.\n"
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("Error writing to temp file: %v", err)
	}
	tmpFile.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	// Run with a start anchor that matches a section.
	exitCode := runMain([]string{"--start", "section-one", tmpFile.Name()}, strings.NewReader("exit\n"), stdout, stderr)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 when starting at anchor, got %d", exitCode)
	}
	got := stdout.String()
	if !strings.Contains(got, "Section One") {
		t.Errorf("Expected output to contain 'Section One', got: %s", got)
	}

	stdout = new(bytes.Buffer)
	stderr = new(bytes.Buffer)
	// Run with a non-existing start anchor.
	exitCode = runMain([]string{"--start", "non-existing", tmpFile.Name()}, strings.NewReader("exit\n"), stdout, stderr)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 when start anchor does not exist, got %d", exitCode)
	}
	if strings.Contains(stdout.String(), "Section One") {
		t.Errorf("Expected output to not contain 'Section One', got: %s", stdout.String())
	}
}
