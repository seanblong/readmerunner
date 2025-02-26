package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/seanblong/readmerunner"
)

// DefaultPrompt reads a line from the provided reader after printing msg.
// This is primarily for testing purposes to mock user input.
func defaultPrompt(r *bufio.Reader, w io.Writer, msg string) string {
	fmt.Fprint(w, msg)
	input, _ := r.ReadString('\n')
	return strings.TrimSpace(input)
}

func runMain(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	var (
		tocFlag     bool
		startAnchor string
		logFile     string
	)

	// Create a new flag set so tests can supply arguments.
	fs := flag.NewFlagSet("readme-runner", flag.ContinueOnError)
	// Redirect error output from fs to stderr.
	fs.SetOutput(stderr)

	fs.BoolVar(&tocFlag, "toc", false, "Print table of contents")
	fs.StringVar(&startAnchor, "start", "", "Anchor text where to start in run mode")
	fs.StringVar(&logFile, "log", "readme-runner.log", "Path to log file")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(stderr, "Error parsing flags:", err)
		return 1
	}

	if fs.NArg() != 1 {
		fmt.Fprintln(stderr, "Usage: readme-runner [options] <README.md>")
		return 1
	}
	readmePath := fs.Arg(0)
	mdContent, err := os.ReadFile(readmePath)
	if err != nil {
		fmt.Fprintln(stderr, "Error reading file:", err)
		return 1
	}

	// Open the log file for appending. Create it if it doesn't exist.
	logF, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintln(stderr, "Error opening log file:", err)
		return 1
	}
	defer logF.Close()

	// Use a multiwriter to output to both stdout and the log file.
	multiOut := io.MultiWriter(stdout, logF)

	if tocFlag {
		err = readmerunner.PrintTOC(multiOut, mdContent)
		if err != nil {
			fmt.Fprintln(stderr, "Error printing TOC:", err)
			return 1
		}
	} else {
		reader := bufio.NewReader(stdin)
		promptFunc := func(msg string) string {
			return defaultPrompt(reader, stdout, msg)
		}
		err = readmerunner.RunMarkdown(mdContent, startAnchor, multiOut, promptFunc)
		if err != nil {
			log.Println("Error running markdown:", err)
			return 1
		}
	}
	return 0
}

func main() {
	os.Exit(runMain(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
