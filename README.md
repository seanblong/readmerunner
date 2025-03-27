# Readme Runner

[![Coverage](https://img.shields.io/badge/Coverage-84.6%25-brightgreen)](https://github.com/seanblong/readmerunner/actions/workflows/test.yaml)
[![CI](https://github.com/seanblong/embedmd/actions/workflows/test.yaml/badge.svg)](https://github.com/seanblong/readmerunner/actions/workflows/test.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/seanblong/readmerunner)](https://goreportcard.com/report/github.com/seanblong/readmerunner)
[![pre-commit.ci status](https://results.pre-commit.ci/badge/github/seanblong/readmerunner/main.svg)](https://results.pre-commit.ci/latest/github/seanblong/readmerunner/main)

Readme Runner is a tool to traverse a README file one section at a time and optionally
execute code snippets.  It is designed to simplify executing runbooks and workflows
and encourage better documentation.

## Installing

### From Source

```bash
go install github.com/seanblong/readmerunner/readme-runner@latest
```

### From Binary

Download the latest binary from the [releases page][2].

## Building

```bash
go build -o readme-runner .
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
  -tags string
          Tags to run (comma-separated)
  -toc
        Print table of contents
```

### Examples

Basic execution:

```console
❯ ./readme-runner ./README.md

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

Printing the table of contents:

```console
❯ ./readme-runner -toc ./README.md
- Readme Runner (readme-runner)
  - Installing (installing)
    - From Source (from-source)
    - From Binary (from-binary)
  - Building (building)
  - Usage (usage)
    - Full Usage (full-usage)
    - Example (example)
```

Running from a specific section:

```console
❯ ./readme-runner -start from-binary ./README.md

### From Binary

Download the latest binary from the [releases page][2].

> Press Enter to continue to [Building] (or type 'exit'):
```

## Environment Variables

The Readme Runner can read environment variables from you machine for use within
the code snippets.  This can be useful for setting up credentials or other
configurations.

Similarly, the code snippets themselves can define the env variables and be passed
onto later code.  This can be useful for defining configurations as groups and
skipping the ones not needed (see example below).

### Example: Using Variables Between Snippets

You can run this example yourself by running,

```console
./readme-runner --start example-using-variables-between-snippets ./README.md
```

```bash
foo=bar
```

```bash
foo=baz
```

```bash
echo $foo
```

## Prompts

If a user prompt is needed, then these can't be executed within the subshell.  Instead,
Readme Runner offers an alternative mechanism to request user input.  This is done
by leveraging a hidden markdown command, `prompt`.

Within your README file you can include lines like this that will not be seen in
the rendered document.

```markdown
[prompt]:# (name "message" [options] default)`
```

### Example: Using Prompts

[prompt]:# (foo "Hello world!" [y] n)

You can run this example yourself by running,

```console
./readme-runner --start example-using-prompts ./README.md
```

## Tags

Tags can be used to help run specific sections of the document or to make sure
certain sections are always run, such as setting environment variables, when using
the `-start` flag to skip ahead.

To add tags simply add a `[tags]:# (tag1 tag2)` line to the section you want to
tag, just below the header, e.g.,

```markdown
## Section

[tags]:# (tag1 tag2)
```

To run this section, you would use the `-tags` flag, e.g.,

```console
./readme-runner --tags tag1 ./README.md
```

or

```console
./readme-runner --tags tag2 ./README.md
```

or

```console
./readme-runner --tags tag1,tag2 ./README.md
```

However, supplying a tag that does not match this section, e.g., `tag3`, would skip
the section.

There's also a special tag, `always`, that will always run the section.  Sections
tagged `always` will run even if a different tag is supplied and will run even when
using the `-start` flag ahead of the section.

## Verification

Readme Runner provides a reserved way to execute verification steps.  Simply add
a code fence with the language `verify` and the code snippet will be executed in
a subshell.  The exit code of the subshell will determine if the verification
passed or failed.  This can be used to wait for processes to complete, check
environment variables, or any other verification step.

For example, the following code snippet will always pass,

````markdown
```verify
return 0
```
````

Here we can use `grep` to detect if a file exists,

````markdown
```verify
ls | grep -q README.md
```
````

If it does we expect the exit code to be 0, if not we expect the return/exit code
to be 1.  The verify step will prompt to rerun the verification step if it fails.


<!-- links -->
[1]: https://gist.github.com/asabaylus/3071099
[2]: https://github.com/seanblong/readmerunner/releases
