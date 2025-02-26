# Readme Runner

[![CI](https://github.com/seanblong/embedmd/actions/workflows/test.yaml/badge.svg)](https://github.com/seanblong/readmerunner/actions/workflows/test.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/seanblong/readmerunner)](https://goreportcard.com/report/github.com/seanblong/readmerunner)
[![pre-commit.ci status](https://results.pre-commit.ci/badge/github/seanblong/readmerunner/main.svg)](https://results.pre-commit.ci/latest/github/seanblong/readmerunner/main)

Readme Runner is a tool to traverse a README file section by section and optionally
execute code snippets.  It is designed to make executing runbooks and workflows
easier as well as encourage better documentation.

## Installing

### From Source

```bash
go install github.com/seanblong/embedmd@latest
```

### From Binary

Download the latest binary from the [releases page](https://github.com/seanblong/readmerunner).

## Building

```bash
go build -o readmerunner ./cmd/main.go
```

## Usage

The default behavior of Readme Runner is to print the README file to the console
one section at a time.  After each section, the user is prompted to continue to
the next section.  When the runner encounters a code snippet, it can execute the
code and print the output to the console.  The user can also choose to skip the
code snippet and continue to the next section.

The output is logged to a file, `readme-runner.log`, by default.  This log file
can be helpful to track the progress of the run and to see the output of the code
snippets.

You can skip to a specific section by using the `--start` flag.  This flag takes
a [Markdown Anchor][1] as an argument.

### Full Usage

```bash
❯ ./readmerunner -h
Usage: readme-runner [options] <README.md>
  -log string
        Path to log file (default "readme-runner.log")
  -start string
        Anchor text where to start in run mode
  -toc
        Print table of contents
```

### Example

```console
❯ ./readmerunner ./README.md

# Readme Runner

Readme Runner is a tool to traverse a README file section by section and optionallyexecute code snippets.  It is designed to make executing runbooks and workflowseasier as well as encourage better documentation.

> Press Enter to continue to [Installing] (or type 'exit'):

## Installing


> Press Enter to continue to [From Source] (or type 'exit'):

### From Source


\`\`\`bash
go install github.com/seanblong/embedmd@latest
\`\`\`

> Run code? (r=run, s=skip, x=exit) [default s]:

> Press Enter to continue to [From Binary] (or type 'exit'):
```

<!-- links -->
[1]: https://gist.github.com/asabaylus/3071099