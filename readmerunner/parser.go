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
	Tags  []string
}

// getHeadingText extracts the text from a header line and prints the header
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
	var b strings.Builder
	for _, r := range lower {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '-' {
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
// or prompt directives (lines starting with "[prompt]:#").
func parseSections(mdContent []byte, start string, userTags []string) []Section {
	var sections []Section
	scanner := bufio.NewScanner(strings.NewReader(string(mdContent)))
	current := Section{Type: SectionText, Lines: []string{}}
	pendingTags := []string{}
	inCodeBlock := false
	codeFence := "```"

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Check for a tags directive.
		if strings.HasPrefix(trimmed, "[tags]:#") {
			if tags, err := parseTags(trimmed); err == nil {
				pendingTags = append(pendingTags, tags...)
				current.Tags = pendingTags
			}
			continue
		}

		// If in a code block, accumulate lines.
		if inCodeBlock {
			current.Lines = append(current.Lines, line)
			if strings.HasPrefix(trimmed, codeFence) {
				inCodeBlock = false
				sections = append(sections, current)
				current = Section{Type: SectionText, Lines: []string{}, Tags: pendingTags}
			}
			continue
		}

		// Start of a code block.
		if strings.HasPrefix(trimmed, codeFence) {
			if len(current.Lines) > 0 {
				sections = append(sections, current)
			}
			current = Section{Type: SectionCode, Lines: []string{}, Tags: pendingTags}
			current.Lines = append(current.Lines, line)
			inCodeBlock = true
			continue
		}

		// A header line starts with "#"
		if strings.HasPrefix(trimmed, "#") {
			if len(current.Lines) > 0 {
				sections = append(sections, current)
			}
			current = Section{Type: SectionHeader, Lines: []string{}, Tags: pendingTags}
			current.Lines = append(current.Lines, line)
			pendingTags = nil
			continue
		}

		// A parameter/prompt directive.
		if strings.HasPrefix(trimmed, "[prompt]:#") {
			if len(current.Lines) > 0 {
				sections = append(sections, current)
			}
			sections = append(sections, Section{Type: SectionPrompt, Lines: []string{line}, Tags: pendingTags})
			current = Section{Type: SectionText, Lines: []string{}}
			continue
		}

		// Otherwise, treat as normal text.
		current.Lines = append(current.Lines, line)
	}
	if len(current.Lines) > 0 {
		sections = append(sections, current)
	}

	started := start == ""
	filtered := []Section{}
	for _, sec := range sections {
		if !started && sec.Type == SectionHeader {
			header, _ := getHeadingText(sec.Lines[0])
			if normalizeAnchor(header) == start {
				started = true
			}
		}
		if checkForAlwaysTag(sec.Tags) {
			filtered = append(filtered, sec)
			continue
		}
		if started {
			if len(userTags) > 0 {
				if checkSectionTag(sec.Tags, userTags) {
					filtered = append(filtered, sec)
				}
			} else {
				filtered = append(filtered, sec)
			}
		}
	}

	if started {
		return filtered
	} else {
		return nil
	}
}

func printLines(w io.Writer, lines []string) {
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
}

func processCodeBlock(w io.Writer, promptFunc func(string) string, code []string, choice string) (err error, exit bool) {
	// Empty code block, just print it.
	if len(code) <= 2 {
		printLines(w, code)
		return nil, false
	}
	// Check the language of the code block.
	// The first line should be the fence with the language.
	var language string
	parts := strings.Split(code[0], "```")
	if len(parts) > 1 {
		language = parts[1]
	}
	codeText := strings.Join(code[1:len(code)-1], "\n")
	runner := GetRunner(language)

	if choice == "" {
		if runner == nil {
			strings.ToLower(strings.TrimSpace(promptFunc("\n> No runner for this language or missing code fence language. Press Enter to continue: ")))
			return nil, false
		} else {
			choice = strings.ToLower(strings.TrimSpace(promptFunc("\n> Run code? (r=run, s=skip, x=exit) [default s]: ")))
		}
	}
	switch choice {
	case "r":
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
			err, exit := processCodeBlock(w, promptFunc, code, "r")
			if err != nil {
				return err, exit
			}
		case "x":
			return nil, true
		case "s", "":
			return nil, false
		default:
			err, exit := processCodeBlock(w, promptFunc, code, "r")
			if err != nil {
				return err, exit
			}
		}
	case "x":
		return nil, true
	case "s", "":
		return nil, false
	default:
		err, exit := processCodeBlock(w, promptFunc, code, "")
		if err != nil {
			return err, exit
		}
	}
	return nil, false
}

// PrintTOC parses the markdown content and writes a table-of-contents.
func PrintTOC(w io.Writer, mdContent []byte) error {
	sections := parseSections(mdContent, "", nil)
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
func RunMarkdown(mdContent []byte, startAnchor string, tags []string, w io.Writer, promptFunc func(string) string) error {
	sections := parseSections(mdContent, startAnchor, tags)
	for i, sec := range sections {
		switch sec.Type {
		case SectionCode:
			fmt.Fprintln(w, strings.Join(sec.Lines, "\n"))
			err, exit := processCodeBlock(w, promptFunc, sec.Lines, "")
			if err != nil {
				return err
			}

			if exit {
				return nil
			}
			continue
		case SectionPrompt:
			for ok := false; !ok; {
				kv, err := processPrompt(promptFunc, sec.Lines)
				if err != nil {
					fmt.Fprintln(w, err)
					continue
				} else {
					fmt.Fprintln(w)
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
