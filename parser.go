package readmerunner

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"unicode"
)

// SectionType indicates what kind of section we have.
type SectionType int

const (
	SectionText SectionType = iota
	SectionHeader
	SectionCode
	SectionPrompt
	SectionUnknown
)

// Section holds a slice of lines and a type.
type Section struct {
	Type  SectionType
	Lines []string
}

// getHeadingTextSimple extracts the text from a header line and prints the header
// level (number of leading #s).
func getHeadingText(header string) (string, int) {
	// Remove all leading #s and trim whitespace.
	clean := strings.TrimSpace(strings.TrimLeft(header, "#"))
	// Count the number of leading #s.
	level := 0
	for _, r := range header {
		if r == '#' {
			level++
		} else {
			break
		}
	}
	return clean, level
}

// normalizeAnchor converts a header string into a markdown anchor.
// It converts the text to lowercase, removes non-alphanumeric characters (except spaces),
// and replaces spaces with dashes
func normalizeAnchor(header string) string {
	lower := strings.ToLower(header)
	// Remove non-alphanumeric characters (allow spaces).
	var b strings.Builder
	for _, r := range lower {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
			b.WriteRune(r)
		}
	}
	anchor := strings.ReplaceAll(b.String(), " ", "-")
	// Optionally, collapse multiple dashes (if needed).
	re := regexp.MustCompile("-+")
	anchor = re.ReplaceAllString(anchor, "-")

	return anchor
}

// parseSections reads the markdown content line‐by‐line and splits it into sections.
// Sections are delimited by header lines (starting with "#"), code block delimiters (```),
// or parameter directives (lines starting with "[prompt]:#").
func parseSections(mdContent []byte) []Section {
	var sections []Section
	scanner := bufio.NewScanner(strings.NewReader(string(mdContent)))
	current := Section{Type: SectionText, Lines: []string{}}
	inCodeBlock := false
	codeFence := "```"

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// If we're in a code block, check for the closing fence.
		if inCodeBlock {
			current.Lines = append(current.Lines, line)
			if strings.HasPrefix(trimmed, codeFence) {
				// End of code block.
				sections = append(sections, current)
				current = Section{Type: SectionText, Lines: []string{}}
				inCodeBlock = false
			}
			continue
		}

		// Check for start of a code block.
		if strings.HasPrefix(trimmed, codeFence) {
			// Flush current section if any.
			if len(current.Lines) > 0 {
				sections = append(sections, current)
			}
			current = Section{Type: SectionCode, Lines: []string{line}}
			inCodeBlock = true
			continue
		}

		// Check for a header line.
		if strings.HasPrefix(trimmed, "#") {
			// New header: flush current section.
			if len(current.Lines) > 0 {
				sections = append(sections, current)
			}
			current = Section{Type: SectionHeader, Lines: []string{line}}
			// (Headers usually stand alone; they may be followed by text that becomes a text section.)
			continue
		}

		// Check for a parameter/prompt directive.
		if strings.HasPrefix(trimmed, "[prompt]:#") || strings.HasPrefix(trimmed, "[param]:#") {
			// Flush current section.
			if len(current.Lines) > 0 {
				sections = append(sections, current)
			}
			// The prompt directive is assumed to be on a single line.
			sections = append(sections, Section{Type: SectionPrompt, Lines: []string{line}})
			// Start a new text section.
			current = Section{Type: SectionText, Lines: []string{}}
			continue
		}

		// Otherwise, add line to current section.
		current.Lines = append(current.Lines, line)
	}

	// If in a code block at EOF then switch it to text.
	if inCodeBlock {
		current.Type = SectionText
	}

	if len(current.Lines) > 0 {
		sections = append(sections, current)
	}
	return sections
}

func printLines(w io.Writer, lines []string) {
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
}

func processCodeBlock(w io.Writer, promptFunc func(string) string, code []string, choice string) error {
	// Empty code block, just print it.
	if len(code) <= 2 {
		printLines(w, code)
		return nil
	}
	// Check the language of the code block.
	// The first line should be the fence with the language.
	var language string
	parts := strings.Split(code[0], "```")
	if len(parts) > 1 {
		language = parts[1]
	}
	codeText := strings.Join(code[1:len(code)-1], "\n")
	if choice == "" {
		choice = strings.ToLower(strings.TrimSpace(promptFunc("\n> Run code? (r=run, s=skip, x=exit) [default s]: ")))
	}
	switch choice {
	case "r":
		runner := GetRunner(language)
		if runner == nil {
			fmt.Fprintf(w, "No runner for language: %s\n", language)
			return nil
		}
		out, err := runner.Run(codeText)
		if err != nil {
			fmt.Fprintf(w, "\n> Error: %s", err.Error())
		}
		if out == "" {
			out = "(no output)\n"
		}
		fmt.Fprintf(w, "\n> Output: %s", out)

		// Prompt after execution: continue, rerun, or exit.
		nextChoice := strings.ToLower(strings.TrimSpace(promptFunc("\n> Continue? (r=rerun, s=continue, x=exit) [default s]: ")))
		switch nextChoice {
		case "r":
			err := processCodeBlock(w, promptFunc, code, "r")
			if err != nil {
				return err
			}
		case "x":
			break
		default:
			// For any other input (including "c"), break the loop to continue to the next section.
			break
		}
	case "x":
		// Use a special error value to signal exit.
		return nil
	}
	return nil
}

// PrintTOC parses the markdown content and writes a table-of-contents.
func PrintTOC(w io.Writer, mdContent []byte) error {
	sections := parseSections(mdContent)
	for _, sec := range sections {
		if sec.Type == SectionHeader {
			// Get the anchor text.
			header, level := getHeadingText(sec.Lines[0])
			// Normalize the anchor.
			anchor := normalizeAnchor(header)
			indent := strings.Repeat("  ", level-1)
			fmt.Fprintf(w, "%s- %s (%s)\n", indent, header, anchor)
		}
	}
	return nil
}

// RunMarkdownSimple processes the markdown content (without using Goldmark)
// and prints sections until a delimiter is reached, then prompts the user.
func RunMarkdown(mdContent []byte, startAnchor string, w io.Writer, promptFunc func(string) string) error {
	sections := parseSections(mdContent)
	started := startAnchor == ""
	for i, sec := range sections {
		if !started && sec.Type == SectionHeader {
			header, _ := getHeadingText(sec.Lines[0])

			started = startAnchor == normalizeAnchor(header)
		}

		if !started {
			continue
		}

		switch sec.Type {
		case SectionCode:
			fmt.Fprintln(w, strings.Join(sec.Lines, "\n"))
			err := processCodeBlock(w, promptFunc, sec.Lines, "")
			if err != nil {
				return err
			}
			continue
		case SectionPrompt:
			for ok := false; !ok; {
				kv, err := processPrompt(promptFunc, sec.Lines)
				if err != nil {
					fmt.Fprintln(w, err)
					continue
				} else {
					ok = true
				}
				for k, v := range kv {
					os.Setenv(k, v)
				}
			}
			continue
		case SectionHeader:
			fmt.Fprintln(w, strings.Join(sec.Lines, "\n"))
			if i < len(sections)-1 {
				nextSection := sections[i+1]
				if nextSection.Type == SectionHeader {
					// If the next section is a header, get its text.
					heading := nextSection.Lines[0]
					nextHeaderText, _ := getHeadingText(heading)
					promptMsg := fmt.Sprintf("\n> Press Enter to continue to [%s] (or type 'exit'): ", nextHeaderText)
					if strings.ToLower(promptFunc(promptMsg)) == "exit" {
						return nil
					} else {
						fmt.Fprintln(w)
					}
				} else {
					continue
				}
			}
		case SectionText:
			fmt.Fprintln(w, strings.Join(sec.Lines, "\n"))
		}
	}
	fmt.Fprintln(w, "\n> README complete!")
	return nil
}
