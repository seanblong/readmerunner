package readmerunner

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"unicode"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// CodeRunner defines a standard interface to run code snippets.
type CodeRunner interface {
	Run(code string) (string, error)
}

// BashRunner implements CodeRunner for bash.
type BashRunner struct{}

// Run executes the given bash code via the shell.
func (r *BashRunner) Run(code string) (string, error) {
	cmd := exec.Command("bash", "-c", code)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// GetRunner returns a CodeRunner based on the provided language.
// For now only "bash" is supported, but this can be extended, e.g. Python, Ruby.
// Fences without a language will be ignored.
func GetRunner(lang string) CodeRunner {
	switch lang {
	case "bash":
		return &BashRunner{}
	default:
		return nil
	}
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

// PrintTOC parses the markdown content and writes a table-of-contents to w.
func PrintTOC(w io.Writer, mdContent []byte) error {
	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(mdContent))
	return ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if n.Kind() == ast.KindHeading && entering {
			heading := n.(*ast.Heading)
			indent := strings.Repeat("  ", heading.Level-1)
			var headerText strings.Builder
			for c := n.FirstChild(); c != nil; c = c.NextSibling() {
				if textNode, ok := c.(*ast.Text); ok {
					headerText.Write(textNode.Segment.Value(mdContent))
				}
			}
			anchor := normalizeAnchor(headerText.String())
			fmt.Fprintf(w, "%s- %s (%s)\n", indent, headerText.String(), anchor)
		}
		return ast.WalkContinue, nil
	})
}

// DefaultPrompt reads a line from the provided reader after printing msg.
// This is primarily for testing purposes to mock user input.
func DefaultPrompt(reader *bufio.Reader, msg string) string {
	fmt.Print(msg)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// renderHeader prints a markdown header with its level and text.
func renderHeader(w io.Writer, n *ast.Heading, mdContent []byte) {
	var headerText strings.Builder
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if textNode, ok := c.(*ast.Text); ok {
			headerText.Write(textNode.Segment.Value(mdContent))
		}
	}
	hashes := strings.Repeat("#", n.Level)
	fmt.Fprintf(w, "\n%s %s\n", hashes, headerText.String())
}

func renderBaseContent(w io.Writer, n ast.Node, mdContent []byte) {
	switch n.Kind() {
	case ast.KindText:
		textNode := n.(*ast.Text)
		fmt.Fprint(w, string(textNode.Segment.Value(mdContent)))
	default:
		if n.Lines() != nil && n.Lines().Len() > 0 {
			for i := 0; i < n.Lines().Len(); i++ {
				line := n.Lines().At(i)
				fmt.Fprint(w, string(line.Value(mdContent)))
			}
		}
	}
}

// renderNodeContent recursively prints the content of a generic node, e.g, text
func renderNodeContent(w io.Writer, n ast.Node, mdContent []byte) {
	switch n.Kind() {
	case ast.KindListItem:
		listItem := n.(*ast.ListItem)
		if listItem.HasChildren() {
			for c := listItem.FirstChild(); c != nil; c = c.NextSibling() {
				renderBaseContent(w, c, mdContent)
			}
		}
	case ast.KindBlockquote:
		// renderBlockquote handles block quote nodes by rendering each child and prefixing each line with "> ".
		bq := n.(*ast.Blockquote)
		// Iterate over children of the blockquote.
		fmt.Fprintln(w)
		for c := bq.FirstChild(); c != nil; c = c.NextSibling() {
			// Capture the rendered content in a temporary buffer.
			var buf strings.Builder
			renderBaseContent(&buf, c, mdContent)
			// Split the content by newlines.
			lines := strings.Split(buf.String(), "\n")
			for _, line := range lines {
				// Print each non-empty line with a "> " prefix.
				if strings.TrimSpace(line) != "" {
					fmt.Fprintf(w, "> %s\n", line)
				} else {
					fmt.Fprintln(w, ">")
				}
			}
		}
	case ast.KindParagraph:
		fmt.Fprintln(w)
		paragraph := n.(*ast.Paragraph)
		if paragraph.HasChildren() {
			for c := paragraph.FirstChild(); c != nil; c = c.NextSibling() {
				renderBaseContent(w, c, mdContent)
			}
		}
		fmt.Fprintln(w)
	default:
		return
	}
}

// renderList handles both ordered and unordered lists by iterating over each list item.
// If a list item consists solely of a single paragraph, its inline text is rendered on the same line.
func renderList(w io.Writer, list *ast.List, mdContent []byte, indent string) {
	index := 1
	fmt.Fprintln(w)
	for li := list.FirstChild(); li != nil; li = li.NextSibling() {
		if li.Kind() == ast.KindListItem {
			var bullet string
			if list.IsOrdered() {
				bullet = fmt.Sprintf("%d.", index)
				index++
			} else {
				bullet = "-"
			}
			// Print bullet with provided indent.
			fmt.Fprintf(w, "%s%s ", indent, bullet)
			renderNodeContent(w, li, mdContent)

			fmt.Fprintln(w)
		}
	}
}

// RunMarkdown iterates through the markdown content grouped by section.
// A section is defined as a header and all content until the next header.
// It writes the section’s content to w and uses promptFunc to get user input.
// Prompts are printed with newlines before and after.
// Code blocks are handled interactively (prompting for run/skip/exit).
// Instead of calling os.Exit, it stops processing when the user types "exit".
func RunMarkdown(mdContent []byte, startAnchor string, w io.Writer, promptFunc func(string) string) error {
	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(mdContent))

	// Group nodes into sections (each section is a slice of ast.Node)
	var sections [][]ast.Node
	var currentSection []ast.Node

	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		if n.Kind() == ast.KindHeading {
			if len(currentSection) > 0 {
				sections = append(sections, currentSection)
			}
			currentSection = []ast.Node{n}
		} else {
			// Content before any heading becomes its own section.
			if currentSection == nil {
				currentSection = []ast.Node{n}
			} else {
				currentSection = append(currentSection, n)
			}
		}
	}
	if len(currentSection) > 0 {
		sections = append(sections, currentSection)
	}

	// If a start anchor is provided, skip sections until it is found.
	started := startAnchor == ""
	for i, section := range sections {
		// If the section starts with a heading, check for the start anchor.
		if len(section) > 0 && section[0].Kind() == ast.KindHeading {
			heading := section[0].(*ast.Heading)
			var headerText strings.Builder
			for c := heading.FirstChild(); c != nil; c = c.NextSibling() {
				if textNode, ok := c.(*ast.Text); ok {
					headerText.Write(textNode.Segment.Value(mdContent))
				}
			}
			normalized := normalizeAnchor(headerText.String())
			if !started && normalized == startAnchor {
				started = true
			}
		}
		if !started {
			continue
		}

		// Process the entire section.
		for _, n := range section {
			switch n.Kind() {
			case ast.KindHeading:
				heading := n.(*ast.Heading)
				renderHeader(w, heading, mdContent)
			case ast.KindFencedCodeBlock:
				codeBlock := n.(*ast.FencedCodeBlock)
				language := string(codeBlock.Language(mdContent))

				var codeText strings.Builder
				for i := 0; i < codeBlock.Lines().Len(); i++ {
					line := codeBlock.Lines().At(i)
					codeText.Write(line.Value(mdContent))
				}
				// Print the code block with markdown formatting.
				fmt.Fprintf(w, "\n```%s\n%s```\n", language, codeText.String())
				// Prompt the user immediately after printing the code block.
				choice := promptFunc("\n> Run code? (r=run, s=skip, x=exit) [default s]: ")
				choice = strings.ToLower(strings.TrimSpace(choice))
				switch choice {
				case "r":
					runner := GetRunner(language)
					if runner == nil {
						fmt.Fprintf(w, "No runner for language: %s\n", language)
						continue
					}
					out, err := runner.Run(codeText.String())
					if err != nil {
						fmt.Fprintf(w, "\n> Error: %s", err.Error())
					}
					fmt.Fprintf(w, "\n> Output: %s", out)
				case "x":
					return nil
					// default: skip code block execution.
				}
			case ast.KindList:
				// Render lists using the helper function.
				list := n.(*ast.List)
				renderList(w, list, mdContent, "")
			default:
				renderNodeContent(w, n, mdContent)
			}
		}

		// Default prompt heading, ideally prompt with next section heading.
		var promptMsg = "\n> Press Enter to continue (or type 'exit'): "
		if i < len(sections)-1 {
			nextSection := sections[i+1]
			if len(nextSection) > 0 && nextSection[0].Kind() == ast.KindHeading {
				heading := nextSection[0].(*ast.Heading)
				var nextHeaderText strings.Builder
				for c := heading.FirstChild(); c != nil; c = c.NextSibling() {
					if textNode, ok := c.(*ast.Text); ok {
						nextHeaderText.Write(textNode.Segment.Value(mdContent))
					}
				}
				promptMsg = fmt.Sprintf("\n> Press Enter to continue to [%s] (or type 'exit'): ", nextHeaderText.String())
			}
		}

		if strings.ToLower(promptFunc(promptMsg)) == "exit" {
			return nil
		}
		// If this is the last section and user did not exit, print final message.
		if i == len(sections)-1 {
			fmt.Fprintln(w, "\n> README complete!")
		}
	}
	return nil
}
