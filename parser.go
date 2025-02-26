package readmerunner

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

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

// renderNodeContent recursively prints the content of a non-header node, e.g.,
// paragraph, listItem, blockquote, etc.
func renderNodeContent(w io.Writer, n ast.Node, mdContent []byte) {
	switch n.Kind() {
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
		var bullet string
		if list.IsOrdered() {
			bullet = fmt.Sprintf("%d.", index)
			index++
		} else {
			bullet = "-"
		}
		// Print bullet with provided indent.
		fmt.Fprintf(w, "%s%s ", indent, bullet)
		listItem := li.(*ast.ListItem)
		if listItem.HasChildren() {
			for c := listItem.FirstChild(); c != nil; c = c.NextSibling() {
				// Check for a sublist
				if c.Kind() == ast.KindList {
					renderList(w, c.(*ast.List), mdContent, indent+"  ")
				}
				renderBaseContent(w, c, mdContent)
			}
		}

		fmt.Fprintln(w)
	}
}

// groupSections splits the document into sections.
// Each section is a slice of ast.Node that starts with a heading (or content before any heading).
func groupSections(doc ast.Node) [][]ast.Node {
	var sections [][]ast.Node
	var currentSection []ast.Node

	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		if n.Kind() == ast.KindHeading {
			if len(currentSection) > 0 {
				sections = append(sections, currentSection)
			}
			currentSection = []ast.Node{n}
		} else {
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
	return sections
}

// getHeadingText extracts the text content from a heading node.
func getHeadingText(heading *ast.Heading, mdContent []byte) string {
	var sb strings.Builder
	for c := heading.FirstChild(); c != nil; c = c.NextSibling() {
		if textNode, ok := c.(*ast.Text); ok {
			sb.Write(textNode.Segment.Value(mdContent))
		}
	}
	return sb.String()
}

// processCodeBlock handles a fenced code block: prints it, prompts the user, and runs it if requested.
func processCodeBlock(w io.Writer, codeBlock *ast.FencedCodeBlock, mdContent []byte, promptFunc func(string) string) error {
	language := string(codeBlock.Language(mdContent))
	var codeText strings.Builder
	for i := 0; i < codeBlock.Lines().Len(); i++ {
		line := codeBlock.Lines().At(i)
		codeText.Write(line.Value(mdContent))
	}
	fmt.Fprintf(w, "\n```%s\n%s```\n", language, codeText.String())
	choice := strings.ToLower(strings.TrimSpace(promptFunc("\n> Run code? (r=run, s=skip, x=exit) [default s]: ")))
	switch choice {
	case "r":
		runner := GetRunner(language)
		if runner == nil {
			fmt.Fprintf(w, "No runner for language: %s\n", language)
			return nil
		}
		out, err := runner.Run(codeText.String())
		if err != nil {
			fmt.Fprintf(w, "\n> Error: %s", err.Error())
		}
		if out == "" {
			out = "(no output)\n"
		}
		fmt.Fprintf(w, "\n> Output: %s", out)
	case "x":
		// Use a special error value to signal exit.
		return nil
	}
	return nil
}

// processSection processes all nodes within a single section.
func processSection(section []ast.Node, mdContent []byte, w io.Writer, promptFunc func(string) string) error {
	for _, n := range section {
		switch n.Kind() {
		case ast.KindHeading:
			renderHeader(w, n.(*ast.Heading), mdContent)
		case ast.KindFencedCodeBlock:
			if err := processCodeBlock(w, n.(*ast.FencedCodeBlock), mdContent, promptFunc); err != nil {
				if err.Error() == "exit" {
					return err
				}
			}
		case ast.KindList:
			list := n.(*ast.List)
			renderList(w, list, mdContent, "")
		default:
			renderNodeContent(w, n, mdContent)
		}
	}
	return nil
}

// RunMarkdown is the top-level function that iterates through sections, processes each,
// and prompts the user between sections.
func RunMarkdown(mdContent []byte, startAnchor string, w io.Writer, promptFunc func(string) string) error {
	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(mdContent))
	sections := groupSections(doc)

	started := startAnchor == ""
	for i, section := range sections {
		// If the section starts with a heading, check its normalized anchor.
		if len(section) > 0 && section[0].Kind() == ast.KindHeading {
			heading := section[0].(*ast.Heading)
			headerText := getHeadingText(heading, mdContent)
			if !started && normalizeAnchor(headerText) == startAnchor {
				started = true
			}
		}
		if !started {
			continue
		}

		// Process the current section.
		if err := processSection(section, mdContent, w, promptFunc); err != nil {
			if err.Error() == "exit" {
				return nil
			}
			return err
		}

		// Build prompt message for the next section.
		promptMsg := "\n> Press Enter to continue (or type 'exit'): "
		if i < len(sections)-1 {
			nextSection := sections[i+1]
			if len(nextSection) > 0 && nextSection[0].Kind() == ast.KindHeading {
				heading := nextSection[0].(*ast.Heading)
				nextHeaderText := getHeadingText(heading, mdContent)
				promptMsg = fmt.Sprintf("\n> Press Enter to continue to [%s] (or type 'exit'): ", nextHeaderText)
			}
		}
		if strings.ToLower(promptFunc(promptMsg)) == "exit" {
			return nil
		}
		if i == len(sections)-1 {
			fmt.Fprintln(w, "\n> README complete!")
		}
	}
	return nil
}
