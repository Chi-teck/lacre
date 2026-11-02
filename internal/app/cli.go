// Package app is lacre itself. The root main.go is a shim over Main below, and
// carries the command's own documentation, where `go doc` looks for it.
package app

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
)

// usage is the usage text, written by hand rather than by flag.PrintDefaults.
// --help and -h, before or after build, print it to standard output; a
// command-line error prints it to standard error, after the error line and an
// empty line, and lacre without a command prints it there alone. A Go raw
// string drops carriage returns, so its lines end with LF on every system.
const usage = `lacre turns plain text or an HTML snippet into a full HTML page.

Usage:
  lacre build [options] < input > page.html
  lacre --help
  lacre --version

The build command reads standard input and writes the page to standard
output. The input goes into the page as it is, as HTML.

Options for build:
  --title TEXT   page title (default "Untitled")
  --lang CODE    page language, such as "en" (default: none)
  --head HTML    HTML to add at the end of <head>, as it is; can be
                 given more than once

Environment:
  LACRE_THEME    a theme directory, whose page.gohtml is the page template;
                 when unset or empty, lacre uses its built-in page

Examples:
  echo 'Hello world!' | lacre build > hello-world.html
  lacre build --title 'My page' --lang en < snippet.html > page.html
  lacre build --head "$(cat head.html)" < snippet.html > page.html
`

// terminalHint is what lacre build prints, without the usage, when standard
// input is a terminal.
const terminalHint = `lacre: standard input is a terminal; pipe or redirect the input, for example:
  echo 'Hello world!' | lacre build > hello-world.html
`

// Main is the command: it runs the command line with the process's own
// arguments, LACRE_THEME, standard streams and terminal check, and exits with
// the code run returns. A closed pipe on Unix ends the program before that:
// nothing here catches SIGPIPE, so Go's runtime lets it end the program when a
// write to standard output fails with EPIPE.
func Main() {
	os.Exit(run(os.Args[1:], os.Getenv("LACRE_THEME"), os.Stdin, os.Stdout, os.Stderr, stdinIsTerminal))
}

// run is the command line. It works in this order: it reads the command line,
// opens the theme and parses its page.gohtml, checks that standard input is not
// a terminal, reads all of standard input, runs the template and writes the
// page; the first failure decides the message and the exit code: 0 when the
// page, the usage or the version was written, 1 when reading standard input or
// writing standard output failed, 2 when lacre was called wrongly or the theme
// is broken. themeDir is the value of LACRE_THEME, empty for the built-in
// theme; it is a parameter so that the tests cannot pick up a LACRE_THEME set
// where they run. isTerminal says whether standard input is a terminal: Main
// passes stdinIsTerminal, the tests a fake.
//
// Go's flag package reads the options: first those before the command, then
// those of build. Its output goes to io.Discard, which silences both the
// error and the default usage it would print. -h and -help are its built-in
// flag.ErrHelp, so the first --help ends the reading.
func run(args []string, themeDir string, stdin io.Reader, stdout, stderr io.Writer, isTerminal func() bool) int {
	top := flag.NewFlagSet("lacre", flag.ContinueOnError)
	top.SetOutput(io.Discard)
	showVersion := top.Bool("version", false, "")
	switch err := top.Parse(args); {
	case errors.Is(err, flag.ErrHelp):
		return writeStdout(stdout, stderr, []byte(usage))
	case err != nil:
		return commandLineError(stderr, err.Error())
	case *showVersion:
		// The options before the command have been read; what comes after
		// is not looked at.
		return writeStdout(stdout, stderr, []byte("lacre "+buildVersion()+"\n"))
	case top.NArg() == 0:
		_, _ = io.WriteString(stderr, usage)
		return 2
	case top.Arg(0) != "build":
		return commandLineError(stderr, fmt.Sprintf("unknown command %q", top.Arg(0)))
	}

	// The values go to buildPage as given, which treats an empty one as not
	// given. The usage strings are empty: flag never prints them.
	build := flag.NewFlagSet("build", flag.ContinueOnError)
	build.SetOutput(io.Discard)
	var title, lang string
	var heads []string
	build.StringVar(&title, "title", "", "")
	build.StringVar(&lang, "lang", "", "")
	build.Func("head", "", func(v string) error {
		heads = append(heads, v)
		return nil
	})
	switch err := build.Parse(top.Args()[1:]); {
	case errors.Is(err, flag.ErrHelp):
		return writeStdout(stdout, stderr, []byte(usage))
	case err != nil:
		return commandLineError(stderr, err.Error())
	case build.NArg() > 0:
		return commandLineError(stderr, "build takes no arguments; it reads standard input")
	}

	// A broken theme stops lacre before it reads standard input. The files
	// that include and dataURI read are read later, while the template runs,
	// but still before anything is written.
	th, err := openTheme(themeDir)
	if err != nil {
		return themeError(stderr, themeDir, err)
	}
	defer th.Close()

	if isTerminal() {
		_, _ = io.WriteString(stderr, terminalHint)
		return 2
	}
	// All of it, as it arrives, before anything is written.
	input, err := io.ReadAll(stdin)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "lacre: read standard input: %v\n", stripPath(err))
		return 1
	}
	page, err := buildPage(th, input, title, lang, heads)
	if err != nil {
		return themeError(stderr, themeDir, err)
	}
	return writeStdout(stdout, stderr, page)
}

// commandLineError prints a command-line error to standard error: the line
// "lacre: <msg>", an empty line and the usage. It gives exit code 2.
func commandLineError(stderr io.Writer, msg string) int {
	_, _ = fmt.Fprintf(stderr, "lacre: %s\n\n%s", msg, usage)
	return 2
}

// themeError prints a theme error to standard error: the line
// "lacre: theme <dir>: <err>", or "lacre: built-in theme: <err>" when dir is
// empty, with no usage after it, since the command line was right. A template
// error loses only its "template: " prefix. It gives exit code 2.
//
// An empty dir is the built-in theme, which has no directory to name. Naming
// it keeps whoever changes it from looking for the mistake in LACRE_THEME.
func themeError(stderr io.Writer, dir string, err error) int {
	name := "theme " + dir
	if dir == "" {
		name = "built-in theme"
	}
	_, _ = fmt.Fprintf(stderr, "lacre: %s: %s\n", name, strings.TrimPrefix(err.Error(), "template: "))
	return 2
}

// writeStdout writes b, which is the page, the usage or the version, to
// standard output in one write. If that fails it prints the error and gives
// exit code 1, otherwise 0.
func writeStdout(stdout, stderr io.Writer, b []byte) int {
	if _, err := stdout.Write(b); err != nil {
		_, _ = fmt.Fprintf(stderr, "lacre: write standard output: %v\n", stripPath(err))
		return 1
	}
	return 0
}

// stripPath gives the error under an *fs.PathError. os.Stdin and os.Stdout
// wrap theirs as "read /dev/stdin: …" and "write /dev/stdout: …", on Windows
// too, and the message already names the operation and the stream.
func stripPath(err error) error {
	if pe, ok := errors.AsType[*fs.PathError](err); ok {
		return pe.Err
	}
	return err
}
