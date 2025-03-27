package readmerunner

import (
	"fmt"
	"regexp"
	"strings"
)

// parseTags parses a tag directive of the form:
// [tags]:# (always foo bar)
func parseTags(line string) ([]string, error) {
	re := regexp.MustCompile(`^\[tags\]:#\s*\(\s*([^)]+)\s*\)$`)
	matches := re.FindStringSubmatch(line)
	if matches == nil {
		return nil, fmt.Errorf("invalid tags directive format")
	}
	// Split by whitespace.
	parts := strings.Fields(matches[1])
	return parts, nil
}

func checkForAlwaysTag(tags []string) bool {
	for _, tag := range tags {
		if tag == "always" {
			return true
		}
	}
	return false
}

func checkSectionTag(sectionTags, runTags []string) bool {
	// If runTags is empty, run everything.
	if len(runTags) == 0 {
		return true
	}
	runTags = append(runTags, "always")
	for _, tag := range sectionTags {
		for _, rt := range runTags {
			if tag == rt {
				return true
			}
		}
	}
	return false
}
