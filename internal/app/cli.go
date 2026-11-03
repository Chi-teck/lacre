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
	"path/filepath"
	"strings"
)

// usage is the usage text, written by hand rather than by flag.PrintDefaults.
// --help and -h, before or after a command, print it to standard output; a
// command-line error prints it to standard error, after the error line and an
// empty line, and lacre without a command prints it there alone. A Go raw
// string drops carriage returns, so its lines end with LF on every system.
const usage = `lacre turns plain text or an HTML snippet into a full HTML page.

Usage:
  lacre build [options] < input > page.html
  lacre style-guide [--theme DIR]
  lacre install-skill DIR
  lacre --help
  lacre --version

The build command reads standard input and writes the page to standard
output. The input goes into the page as it is, as HTML.

The style-guide command prints the style guide of the theme in use,
in Markdown, as the theme's author wrote it: what the input can use,
such as a framework and its classes, and what the page already puts
around the input. Read it before writing input for the theme.

The install-skill command writes a skill that teaches AI agents when
and how to use lacre. DIR is the directory that holds an agent's
skills, such as .claude/skills for Claude Code or .agents/skills for
Codex; the skill goes into DIR/lacre/SKILL.md. It creates the
directories it needs and replaces an existing SKILL.md, so running it
again after an upgrade updates the skill. A ~ that the shell did not
expand, as in a quoted '~/.claude/skills', is an error.

Options for build:
  --title TEXT   page title (default "Untitled")
  --lang CODE    page language, such as "en" (default: none)
  --head HTML    HTML to add at the end of <head>, as it is; can be
                 given more than once

Options for build and style-guide:
  --theme DIR    a theme directory, whose page.gohtml is the page
                 template; empty, as in --theme=, for the built-in
                 theme (default: LACRE_THEME)

Environment:
  LACRE_THEME    a theme directory, used without --theme; when unset
                 or empty, lacre uses its built-in theme

Examples:
  echo 'Hello world!' | lacre build > hello-world.html
  lacre build --title 'My page' --lang en < snippet.html > page.html
  lacre build --head "$(cat head.html)" < snippet.html > page.html
  lacre style-guide
  lacre install-skill .claude/skills
  lacre install-skill .agents/skills
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

// run is the command line. For build it works in this order: it reads the
// command line, opens the theme and parses its page.gohtml, checks that
// standard input is not a terminal, reads all of standard input, runs the
// template and writes the page; style-guide opens the theme the same way, then
// reads and writes its style guide (see runStyleGuide); install-skill creates
// the directories it needs and writes the skill (see runInstallSkill). The
// first failure decides the message and the exit code: 0 when the page, the
// style guide, the usage or the version was written or the skill was
// installed, 1 when reading standard input or writing standard output failed
// or, for install-skill, creating a directory or writing the skill failed, 2
// when lacre was called wrongly, the theme is broken or, for style-guide, it
// has no style guide. themeDir is the value of LACRE_THEME, empty for the
// built-in theme, and a --theme replaces it; it is a parameter so that the
// tests cannot pick up a LACRE_THEME set where they run. isTerminal says
// whether standard input is a terminal: Main passes stdinIsTerminal, the tests
// a fake.
//
// Go's flag package reads the options: first those before the command, then
// those of the command. Its output goes to io.Discard, which silences both the
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
	case top.Arg(0) == "style-guide":
		return runStyleGuide(top.Args()[1:], themeDir, stdout, stderr)
	case top.Arg(0) == "install-skill":
		return runInstallSkill(top.Args()[1:], stdout, stderr)
	case top.Arg(0) != "build":
		return commandLineError(stderr, fmt.Sprintf("unknown command %q", top.Arg(0)))
	}

	// The values of --title, --lang and --head go to buildPage as given, which
	// treats an empty one as not given. The usage strings are empty: flag never
	// prints them.
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
	themeOption(build, &themeDir)
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

// themeOption defines --theme in flags, for build and style-guide alike.
// --theme replaces *dir only when it is given, so that it wins over
// LACRE_THEME and an empty one gives the built-in theme, the rule the variable
// follows. The last one wins.
func themeOption(flags *flag.FlagSet, dir *string) {
	flags.Func("theme", "", func(v string) error {
		*dir = v
		return nil
	})
}

// runStyleGuide is the style-guide command, which writes the style guide of
// the theme in use to standard output. args are the arguments after the
// command, and themeDir and the rest are as for run. It opens the theme as
// build does, so that a theme build cannot open fails here too, before anyone
// writes input for it. The template is not run, so the files that include and
// dataURI read are not checked. It does not read standard input, so a terminal
// there is no error.
func runStyleGuide(args []string, themeDir string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("style-guide", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	themeOption(flags, &themeDir)
	switch err := flags.Parse(args); {
	case errors.Is(err, flag.ErrHelp):
		return writeStdout(stdout, stderr, []byte(usage))
	case err != nil:
		return commandLineError(stderr, err.Error())
	case flags.NArg() > 0:
		return commandLineError(stderr, "style-guide takes no arguments; name the theme with --theme")
	}

	th, err := openTheme(themeDir)
	if err != nil {
		return themeError(stderr, themeDir, err)
	}
	defer th.Close()
	guide, err := th.styleGuide()
	if err != nil {
		return themeError(stderr, themeDir, err)
	}
	return writeStdout(stdout, stderr, guide)
}

// runInstallSkill is the install-skill command, which writes the agent skill
// into the skills directory that its one argument names, as lacre/SKILL.md in
// it. args are the arguments after the command; stdout is only for --help. It
// has no options, so --theme is an error too. It does not read standard input,
// so a terminal there is no error.
//
// It creates the directories it needs with mode 0o777 and writes the file with
// 0o666, so that the umask decides, as with mkdir -p and the shell's >. Like >,
// os.WriteFile truncates an existing SKILL.md, which keeps its mode, and
// follows a symbolic link; other files in the lacre directory stay. A
// temporary file renamed into place would survive a failed write, but would
// replace the link and the mode, for a short file written in one call.
//
// filepath.Join cleans the paths and gives them the system's separator, and a
// relative directory stays relative, so the messages name the file as the user
// typed it. An empty directory is an error, since Join would drop it and write
// into the current directory, which . already names.
//
// A directory whose first name is ~, as in ~/.claude/skills, is an error too.
// The shell expands ~ only when it is not quoted, and cmd.exe and PowerShell
// before 7.6 not at all, so such a ~ reaches lacre as it is, and lacre would
// create a directory named ~ in the current directory, where no agent reads
// skills. The check comes after Join, so ./~ is caught too; a name that only
// starts with ~, such as ~skills, is no error.
func runInstallSkill(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("install-skill", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	switch err := flags.Parse(args); {
	case errors.Is(err, flag.ErrHelp):
		return writeStdout(stdout, stderr, []byte(usage))
	case err != nil:
		return commandLineError(stderr, err.Error())
	case flags.NArg() != 1:
		return commandLineError(stderr, "install-skill takes one argument: the skills directory")
	case flags.Arg(0) == "":
		return commandLineError(stderr, "the skills directory for install-skill is empty")
	}

	skillDir := filepath.Join(flags.Arg(0), "lacre")
	if strings.HasPrefix(skillDir, "~"+string(filepath.Separator)) {
		return commandLineError(stderr, "the skills directory for install-skill starts with ~, which the shell "+
			"did not expand; name the home directory instead, such as $HOME or %USERPROFILE%")
	}
	if err := os.MkdirAll(skillDir, 0o777); err != nil {
		// The directory that failed may be above lacre, such as a file in
		// the way, and the *fs.PathError names it.
		name := skillDir
		if pe, ok := errors.AsType[*fs.PathError](err); ok {
			name = pe.Path
		}
		_, _ = fmt.Fprintf(stderr, "lacre: create directory %s: %v\n", name, stripPath(err))
		return 1
	}
	path := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(path, []byte(skillText()), 0o666); err != nil {
		_, _ = fmt.Fprintf(stderr, "lacre: write %s: %v\n", path, stripPath(err))
		return 1
	}
	_, _ = fmt.Fprintf(stderr, "lacre: wrote %s\n", path)
	return 0
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
// error loses only its "template: " prefix. It gives exit code 2. For
// style-guide, a theme without a style guide is a theme error too.
//
// An empty dir is the built-in theme, which has no directory to name. Naming
// it keeps whoever changes it from looking for the mistake in --theme or
// LACRE_THEME.
func themeError(stderr io.Writer, dir string, err error) int {
	name := "theme " + dir
	if dir == "" {
		name = "built-in theme"
	}
	_, _ = fmt.Fprintf(stderr, "lacre: %s: %s\n", name, strings.TrimPrefix(err.Error(), "template: "))
	return 2
}

// writeStdout writes b, which is the page, the style guide, the usage or the
// version, to standard output in one write. If that fails it prints the error
// and gives exit code 1, otherwise 0.
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
// install-skill's messages name the operation and the path in lacre's words.
func stripPath(err error) error {
	if pe, ok := errors.AsType[*fs.PathError](err); ok {
		return pe.Err
	}
	return err
}
