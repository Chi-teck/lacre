# lacre

[![CI](https://github.com/Chi-teck/lacre/actions/workflows/ci.yml/badge.svg)](https://github.com/Chi-teck/lacre/actions/workflows/ci.yml)

## What it's for

lacre turns plain text or an HTML snippet into a full HTML page. A bit of text or a snippet is not a page: to open it in
a browser, share it or print it, someone has to wrap it in the boilerplate of a complete HTML document. lacre does that,
and nothing else.

```bash
echo 'Hello world!' | lacre build > hello-world.html
```

Opening `hello-world.html` in a browser shows "Hello world!" on a complete page.

## Install

Ready-made binaries for Linux, macOS and Windows, on amd64 and arm64, will be on the
[releases page](https://github.com/Chi-teck/lacre/releases) from the first release on. There is no release yet.

With Go 1.26 or later:

```bash
go install github.com/Chi-teck/lacre@latest
```

## Usage

`lacre build` reads all of standard input and writes the page to standard output. Messages go to standard error.

```bash
lacre build --title 'My page' --lang en < snippet.html > page.html
```

The page is a complete HTML document. Its head has the UTF-8 charset, a viewport tag so that the page fits a phone's
screen, a generator tag that names lacre, the title (`Untitled` without `--title`) and a small built-in stylesheet,
which gives a centred column that is easy to read, the system fonts, long lines in `<pre>` that wrap, and images and
videos that shrink to fit the column. Its body is your input, as it is. Nothing lacre adds refers to anything outside
the page, so the page is a single file that works offline, unless what you put in it refers to outside files. The same
input and options always give the same bytes.

Without `--lang` the page declares no language, since a wrong one is worse than none. Setting it helps screen readers
and satisfies accessibility checkers.

`--head` adds HTML at the end of `<head>`, as it is, and can be given more than once. It comes after the built-in
stylesheet, so a `<style>` added this way overrides its rules:

```bash
lacre build --head '<style>body { max-width: 60rem }</style>' < snippet.html > page.html
```

To add a file, let the shell read it:

```bash
lacre build --head "$(cat head.html)" < snippet.html > page.html
```

A `--head` value that refers to outside files, such as a `<link>` to a stylesheet or a `<script src>`, makes the page
depend on them.

## Options

The options of `build`:

| Option         | Value      | Default    | Effect                                                                                       |
|----------------|------------|------------|----------------------------------------------------------------------------------------------|
| `--title TEXT` | any string | `Untitled` | The page title                                                                               |
| `--lang CODE`  | any string | none       | The page language, the `lang` attribute of `<html>`; without it there is no `lang` attribute |
| `--head HTML`  | any string | none       | HTML added at the end of `<head>`, as it is; can be repeated                                 |
| `--help`, `-h` | none       |            | Print the usage to standard output and exit with 0                                           |

Before the command, the only options are `--help` (or `-h`) and `--version`: `lacre --help` prints the usage and
`lacre --version` the version.

## The input is HTML

- The input goes into the page as it is and is not cleaned: a script in the input runs when the page is opened. Use
  lacre only on input you trust. The same goes for `--head` values.
- lacre does not repair broken HTML. A snippet with unclosed tags, or a full document with its own `<html>`, goes in as
  it is, and the browser makes of it what it can.
- Plain text is read as HTML too: its line breaks collapse into one paragraph, and `<` and `&` are read as markup. To
  show it as typed, escape it and wrap it in `<pre>`. In Unix shells:

  ```bash
  { echo '<pre>'; sed -e 's/&/\&amp;/g' -e 's/</\&lt;/g' -e 's/>/\&gt;/g' notes.txt; echo '</pre>'; } \
    | lacre build > notes.html
  ```

## Exit codes

| Code | When                                                                                                               |
|------|--------------------------------------------------------------------------------------------------------------------|
| 0    | The page was written, or the usage or version was printed                                                          |
| 1    | Reading standard input or writing standard output failed                                                           |
| 2    | The tool was called wrongly: no command, an unknown command, a bad option, an argument, or a terminal as input     |

## Windows

lacre works in `cmd.exe`, in PowerShell and in Unix shells such as Git Bash. The examples above are for Unix shells.

In `cmd.exe`, `<` and `>` work as in Unix shells. Quote values with double quotes, and leave the quotes out of `echo`,
or they end up in the page:

```bat
echo Hello world! | lacre build > hello-world.html
lacre build --title "My page" --lang en < snippet.html > page.html
```

The file form of `--head`, `--head "$(cat head.html)"`, is for Unix shells: `cmd.exe` has no command substitution and
would pass `$(cat head.html)` to lacre as text, which then ends up in the page. In `cmd.exe`, give `--head` the HTML
itself, or run lacre in Git Bash.

In PowerShell, `echo 'Hello world!' | lacre build > hello-world.html` works too, but PowerShell has no `<`.

Instead, pipe the file in. This works, but PowerShell may change the line endings and add a line break at the end:

```powershell
Get-Content snippet.html -Raw | lacre build > page.html
```

To pass the bytes exactly, let `cmd.exe` do the redirection:

```powershell
cmd /c "lacre build < snippet.html > page.html"
```

Known limit: Windows PowerShell 5.1 can mangle non-ASCII text in pipes and in `>`. PowerShell 7.4 and later and
`cmd /c` do not. PowerShell before 7.3 also drops double quotes inside arguments passed to programs, which breaks
`--head` values with attributes, such as `--head '<meta name="description" content="A test page" />'`.

The PowerShell details in this section have not been checked on a Windows machine yet.

## Known limits

When standard input is a terminal, `lacre build` does not wait for typed input: it prints a hint and exits with 2. In
Git Bash (mintty), lacre cannot tell that standard input is a terminal, so it waits for input, as `cat` does. Ctrl+D
ends the input. On systems other than Linux, macOS and Windows, such as FreeBSD, lacre does not check at all and waits
the same way.

## Development

The commands below need [Task](https://taskfile.dev); `task` alone lists all the tasks.

```bash
task check              # gofmt, go vet (also for macOS), golangci-lint, go test: the check before a commit
task test:update        # rewrite the golden files, then run the tests again
task build              # CGO_ENABLED=0 go build -o lacre .
task release:snapshot   # build the release archives into dist/, published nowhere
```

`task check` also needs golangci-lint, and `task release:snapshot` needs GoReleaser. The golden files in
`internal/app/testdata` fix the expected output. The binaries of `task release:snapshot` print a snapshot version, such
as `lacre v0.0.0-SNAPSHOT-<commit>` before the first tag.

On every push to a branch and on every pull request, continuous integration (CI) runs the checks of `task check` and the
build on Linux, and the tests and the build on Windows. No job runs on macOS: the macOS build is cross-compiled and
vetted, never run. A `v*` tag runs the same checks, then publishes a release with GoReleaser.

## Licence

MIT. See [LICENSE](LICENSE).
