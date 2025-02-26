package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/seanblong/readmerunner" // adjust the import path as needed
)

// DefaultPrompt reads a line from the provided reader after printing msg.
// This is primarily for testing purposes to mock user input.
func defaultPrompt(reader *bufio.Reader, msg string) string {
	fmt.Print(msg)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func main() {
	var (
		tocFlag     bool
		startAnchor string
		logFile     string
	)
	// Define a boolean flag for table-of-contents.
	flag.BoolVar(&tocFlag, "toc", false, "Print table of contents")
	flag.StringVar(&startAnchor, "start", "", "Anchor text where to start in run mode")
	flag.StringVar(&logFile, "log", "readme-runner.log", "Path to log file")
	flag.Parse()

	if flag.NArg() != 1 {
		log.Println("Usage: readme-runner [options] <README.md>")
		os.Exit(1)
	}
	readmePath := flag.Arg(0)
	mdContent, err := os.ReadFile(readmePath)
	if err != nil {
		log.Println("Error reading file:", err)
		os.Exit(1)
	}

	// Open the log file for appending. Create it if it doesn't exist.
	logF, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Error opening log file:", err)
		os.Exit(1)
	}
	defer logF.Close()

	// Use a multiwriter to output to both stdout and the log file.
	multiOut := io.MultiWriter(os.Stdout, logF)

	if tocFlag {
		err = readmerunner.PrintTOC(multiOut, mdContent)
		if err != nil {
			log.Println("Error printing TOC:", err)
			os.Exit(1)
		}
	} else {
		reader := bufio.NewReader(os.Stdin)
		promptFunc := func(msg string) string {
			return defaultPrompt(reader, msg)
		}
		err = readmerunner.RunMarkdown(mdContent, startAnchor, multiOut, promptFunc)
		if err != nil {
			log.Println("Error running markdown:", err)
			os.Exit(1)
		}
	}
}
