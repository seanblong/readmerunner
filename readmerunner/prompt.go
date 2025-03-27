package readmerunner

import (
	"fmt"
	"regexp"
	"strings"
)

type Prompt struct {
	VarName string   // the variable name to save the value into
	Text    string   // the prompt to display to the user
	Options []string // optional valid options (if provided)
	Default string   // optional default value
}

// parsePrompt parses a single prompt line.
// Example line:
// [prompt]:# (eggs "How many eggs?"  [0,1,2,3,4,5,6] 6)
func parsePrompt(line string) (*Prompt, error) {
	// This regex matches:
	//   Group 1: variable name (alphanumeric and underscore)
	//   Group 2: prompt text inside double quotes
	//   Group 3: optional options list (including square brackets)
	//   Group 4: optional default value (non-space token)
	re := regexp.MustCompile(`^\[prompt\]:#\s*\(\s*(\w+)\s+"([^"]+)"\s*(\[[^\]]*\])?\s*(\S+)?\s*\)$`)
	matches := re.FindStringSubmatch(line)
	if matches == nil || len(matches) < 3 {
		return nil, fmt.Errorf("invalid prompt format: %s", line)
	}
	pd := &Prompt{
		VarName: matches[1],
		Text:    matches[2],
	}
	if len(matches) > 3 && matches[3] != "" {
		// Remove brackets and split by spaces
		optionsStr := strings.Trim(matches[3], "[]")
		opts := strings.Fields(optionsStr)
		// opts := strings.Split(optionsStr, ",")
		for i, opt := range opts {
			opts[i] = strings.TrimSpace(opt)
		}
		pd.Options = opts
	}
	if len(matches) > 4 && matches[4] != "" {
		pd.Default = matches[4]
	}
	return pd, nil
}

// processPrompts scans the markdown content for prompt s,
// prompts the user accordingly, validates responses if options are provided,
// and returns a map of variable names to responses.
func processPrompt(promptFunc func(string) string, prompt []string) (map[string]string, error) {
	varMap := make(map[string]string)
	for _, line := range prompt {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[prompt]:#") {
			pd, err := parsePrompt(line)
			if err != nil {
				return nil, err
			}
			// Build a full prompt message.
			fullPrompt := pd.Text
			if len(pd.Options) > 0 {
				fullPrompt += " (options: " + strings.Join(pd.Options, ", ") + ")"
			}
			if pd.Default != "" {
				fullPrompt += fmt.Sprintf(" [default: %s]", pd.Default)
			}
			fullPrompt += ": "

			response := promptFunc("\n" + fullPrompt)

			// If no response and a default is provided, use default.
			if response == "" && pd.Default != "" {
				response = pd.Default
			}

			// Ensure response is a valid option if options are provided.
			if len(pd.Options) > 0 {
				valid := false
				for _, opt := range pd.Options {
					if response == opt {
						valid = true
						break
					}
				}
				if !valid {
					return nil, fmt.Errorf("invalid response for %s. Must be one of %v", pd.VarName, pd.Options)
				}
			}
			varMap[pd.VarName] = response
		}
	}
	return varMap, nil
}
