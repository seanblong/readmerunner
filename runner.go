package readmerunner

import (
	"bufio"
	"io"
	"log"
	"os/exec"
	"strings"
)

// CodeRunner defines a standard interface to run code snippets.
type CodeRunner interface {
	Run(code string) (string, error)
	Close() error
}

// RunnerIO is a wrapper around exec.Cmd to handle stdin/stdout.
// It allows for running code in a persistent shell.
type runnerIO struct {
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	scanner *bufio.Scanner
}

func newRunnerIO(command string) (*runnerIO, error) {
	cmd := exec.Command(command)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	// Merge stderr into stdout so errors are captured.
	cmd.Stderr = cmd.Stdout
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(stdout)
	return &runnerIO{
		cmd:     cmd,
		stdin:   stdin,
		stdout:  stdout,
		scanner: scanner,
	}, nil
}

// Run executes the provided code in the persistent shell.
func (r *runnerIO) Run(code string) (string, error) {
	marker := "__END_OF_SNIPPET__"
	// Append marker so we know when the output for this snippet is done.
	command := code + "\necho " + marker + "\n"
	if _, err := r.stdin.Write([]byte(command)); err != nil {
		return "", err
	}
	var output strings.Builder
	for r.scanner.Scan() {
		line := r.scanner.Text()
		if line == marker {
			break
		}
		output.WriteString(line + "\n")
	}
	if err := r.scanner.Err(); err != nil {
		return output.String(), err
	}
	return output.String(), nil
}

// Close terminates the shell and cleans up resources.
func (r *runnerIO) Close() error {
	if err := r.stdin.Close(); err != nil {
		return err
	}
	if err := r.cmd.Wait(); err != nil {
		return err
	}
	return nil
}

// BashRunner implements CodeRunner for bash.
type BashRunner struct {
	runnerIO
}

// bashRunner is a singleton instance of BashRunner.
var bashRunner *BashRunner

// NewBashRunner spawns a persistent Bash shell.
func NewBashRunner() (*BashRunner, error) {
	runner, err := newRunnerIO("bash")
	if err != nil {
		return nil, err
	}
	return &BashRunner{*runner}, nil
}

// ShellRunner implements CodeRunner for sh.
type ShellRunner struct {
	runnerIO
}

// shellRunner is a singleton instance of ShellRunner.
var shellRunner *ShellRunner

// NewShellRunner spawns a persistent shell.
func NewShellRunner() (*ShellRunner, error) {
	runner, err := newRunnerIO("sh")
	if err != nil {
		return nil, err
	}
	return &ShellRunner{*runner}, nil
}

// GetRunner returns a CodeRunner based on the provided language.
// For now only "bash" is supported, but this can be extended, e.g. Python, Ruby.
// Fences without a language will be ignored.
func GetRunner(lang string) CodeRunner {
	switch lang {
	case "bash":
		if bashRunner == nil {
			runner, err := NewBashRunner()
			if err != nil {
				log.Printf("Error starting bash runner: %v\n", err)
				return nil
			}
			bashRunner = runner
		}
		return bashRunner
	case "sh", "shell":
		if shellRunner == nil {
			runner, err := NewShellRunner()
			if err != nil {
				log.Printf("Error starting shell runner: %v\n", err)
				return nil
			}
			shellRunner = runner
		}
		return shellRunner
	default:
		return nil
	}
}
