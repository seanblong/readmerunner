package readmerunner

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strconv"
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

// VerifyRunner implements CodeRunner for custom verify functions, i.e. scripts
// should return 0 on success and non-zero on failure.
type VerifyRunner struct {
	runnerIO
}

// shellRunner is a singleton instance of ShellRunner.
var verifyRunner *VerifyRunner

// NewShellRunner spawns a persistent shell.
func NewVerifyRunner() (*VerifyRunner, error) {
	runner, err := newRunnerIO("bash")
	if err != nil {
		return nil, err
	}
	return &VerifyRunner{*runner}, nil
}

// Run executes the provided code in the persistent shell, returning "Success" or
// "Failure" based on the exit code.
func (r *VerifyRunner) Run(code string) (string, error) {
	marker := "__END_OF_SNIPPET__"
	exitMarker := "__EXIT_CODE__"

	// Wrap the snippet code in a function.
	// This override of exit prevents the snippet from terminating the persistent shell.
	wrappedCode := fmt.Sprintf(`function __run_snippet() {
	exit() { return "$@"; }
%s
}
__run_snippet
exitCode=$?
echo %s
echo %s $exitCode
`, code, marker, exitMarker)

	if _, err := r.stdin.Write([]byte(wrappedCode)); err != nil {
		return "", err
	}

	var output strings.Builder
	// Read the snippet output until the marker is encountered.
	for r.scanner.Scan() {
		line := r.scanner.Text()
		if line == marker {
			break
		}
		output.WriteString(line + "\n")
	}

	// The next line should contain the exit code.
	var exitLine string
	if r.scanner.Scan() {
		exitLine = r.scanner.Text()
	}
	parts := strings.Fields(exitLine)
	if len(parts) != 2 || parts[0] != exitMarker {
		return "", fmt.Errorf("failed to parse exit code, got: %s", exitLine)
	}
	exitCode, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid exit code: %s", parts[1])
	}
	if exitCode != 0 {
		return fmt.Sprintf("\033[31mFailure [command exited with status %d]\033[0m\n", exitCode), nil
	}
	return "\033[32mSuccess\033[0m\n", nil
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
	case "verify":
		if verifyRunner == nil {
			runner, err := NewVerifyRunner()
			if err != nil {
				log.Printf("Error starting verify runner: %v\n", err)
				return nil
			}
			verifyRunner = runner
		}
		return verifyRunner
	default:
		return nil
	}
}
