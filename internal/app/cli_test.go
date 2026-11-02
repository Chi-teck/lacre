package app

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"strings"
	"testing"
	"testing/iotest"
)

// TestMain lets T-13 run the real program: started with
// LACRE_TEST_BE_LACRE=1, the test binary is lacre itself, since Main exits
// and never returns.
func TestMain(m *testing.M) {
	if os.Getenv("LACRE_TEST_BE_LACRE") == "1" {
		Main()
	}
	m.Run()
}

// notTerminal is the terminal check of every test that is not about the
// terminal.
func notTerminal() bool { return false }

// lacre calls run with input as standard input and a terminal check that says
// "not a terminal", and gives what it wrote to each stream and its exit code.
func lacre(input string, args ...string) (stdout, stderr string, code int) {
	var out, errOut bytes.Buffer
	code = run(args, strings.NewReader(input), &out, &errOut, notTerminal)
	return out.String(), errOut.String(), code
}

// T-1, the row usage.golden: lacre --help (CLI-18).
func TestUsageGolden(t *testing.T) {
	stdout, _, _ := lacre("", "--help")
	checkGolden(t, "usage", []byte(stdout))
}

// T-5, T-7 and T-8: the input is the body, byte for byte, with exit code 0 and
// nothing on standard error (CLI-9). The suffix pins the body to the input, so
// the byte order mark of T-8 is at its start.
func TestBuildPassThrough(t *testing.T) {
	cases := []struct{ name, input string }{
		{"T-5", "one\r\ntwo\r\n\r\nthree\r\n"},
		{"T-7", "\xff\xfe\x00 ok\n"},
		{"T-8", "\ufeffHello world!\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			stdout, stderr, code := lacre(c.input, "build")
			if code != 0 {
				t.Errorf("the exit code is %d, want 0", code)
			}
			if stderr != "" {
				t.Errorf("standard error is %q, want nothing", stderr)
			}
			if want := "\n<body>\n" + c.input + "</body>\n</html>\n"; !strings.HasSuffix(stdout, want) {
				t.Errorf("the page does not end with %q:\n%q", want, stdout)
			}
		})
	}
}

// T-9, over CLI-17, with the cases T-9 lists and two more: the order of the
// --help and --version checks (CLI-7), and several --head values (CLI-5)
// through the command line.
func TestCommandLine(t *testing.T) {
	usageErr := func(msg string) string { return "lacre: " + msg + "\n\n" + usage }
	page := func(title, lang string, heads ...string) string {
		return string(buildPage([]byte("Hello world!\n"), title, lang, heads))
	}
	versionLine := "lacre " + buildVersion() + "\n"
	const noArgs = "build takes no arguments; it reads standard input"

	cases := []struct {
		args           []string
		stdout, stderr string
		code           int
	}{
		{args: nil, stderr: usage, code: 2},
		{args: []string{"foo"}, stderr: usageErr(`unknown command "foo"`), code: 2},
		{args: []string{"--bogus"}, stderr: usageErr("flag provided but not defined: -bogus"), code: 2},
		{args: []string{"build", "--bogus"}, stderr: usageErr("flag provided but not defined: -bogus"), code: 2},
		{args: []string{"build", "--title"}, stderr: usageErr("flag needs an argument: -title"), code: 2},
		{args: []string{"build", "page.html"}, stderr: usageErr(noArgs), code: 2},
		{args: []string{"--title", "X", "build"}, stderr: usageErr("flag provided but not defined: -title"), code: 2},
		{args: []string{"build", "--", "page.html"}, stderr: usageErr(noArgs), code: 2},
		{args: []string{"--help"}, stdout: usage},
		{args: []string{"-h"}, stdout: usage},
		{args: []string{"build", "--help"}, stdout: usage},
		{args: []string{"build", "-h"}, stdout: usage},
		{args: []string{"--version"}, stdout: versionLine},
		{args: []string{"--version", "build", "--bogus"}, stdout: versionLine},
		{args: []string{"build", "--version"}, stderr: usageErr("flag provided but not defined: -version"), code: 2},
		{args: []string{"build", "--title=X"}, stdout: page("X", "")},
		{args: []string{"build", "--title", "X"}, stdout: page("X", "")},
		{args: []string{"build", "--title", ""}, stdout: page("Untitled", "")},
		{args: []string{"build", "--lang", ""}, stdout: page("", "")},
		{args: []string{"build", "--head"}, stderr: usageErr("flag needs an argument: -head"), code: 2},
		{args: []string{"--version", "--help"}, stdout: usage}, // --help wins (CLI-7)
		{args: []string{"build", "--head", "a", "--head", "b"}, stdout: page("", "", "a", "b")},
	}
	for _, c := range cases {
		stdout, stderr, code := lacre("Hello world!\n", c.args...)
		if stdout != c.stdout {
			t.Errorf("lacre %q: standard output is %q, want %q", c.args, stdout, c.stdout)
		}
		if stderr != c.stderr {
			t.Errorf("lacre %q: standard error is %q, want %q", c.args, stderr, c.stderr)
		}
		if code != c.code {
			t.Errorf("lacre %q: the exit code is %d, want %d", c.args, code, c.code)
		}
	}
}

// T-10: some input arrives and then the read fails. The error is the shape
// os.Stdin gives, so the test also covers the removal of the path.
func TestReadFails(t *testing.T) {
	readErr := &fs.PathError{Op: "read", Path: "/dev/stdin", Err: errors.New("boom")}
	stdin := io.MultiReader(strings.NewReader("Hello world!\n"), iotest.ErrReader(readErr))
	var stdout, stderr bytes.Buffer
	code := run([]string{"build"}, stdin, &stdout, &stderr, notTerminal)
	if code != 1 {
		t.Errorf("the exit code is %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Errorf("standard output is %q, want nothing", stdout.Bytes())
	}
	if want := "lacre: read standard input: boom\n"; stderr.String() != want {
		t.Errorf("standard error is %q, want %q", stderr.Bytes(), want)
	}
}

// failingWriter is a standard output that takes nothing.
type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

// T-11: the page, the usage and the version all go to standard output, and a
// failed write of any of them is an error.
func TestWriteFails(t *testing.T) {
	writeErr := &fs.PathError{Op: "write", Path: "/dev/stdout", Err: errors.New("boom")}
	for _, args := range [][]string{{"build"}, {"--help"}, {"--version"}} {
		var stderr bytes.Buffer
		code := run(args, strings.NewReader("Hello world!\n"), failingWriter{writeErr}, &stderr, notTerminal)
		if code != 1 {
			t.Errorf("lacre %q: the exit code is %d, want 1", args, code)
		}
		if want := "lacre: write standard output: boom\n"; stderr.String() != want {
			t.Errorf("lacre %q: standard error is %q, want %q", args, stderr.Bytes(), want)
		}
	}
}

// T-12: a terminal check that says "terminal". The hint is written out here
// rather than taken from terminalHint, so that the test pins CLI-17's words.
func TestTerminal(t *testing.T) {
	const input = "Hello world!\n"
	const hint = "lacre: standard input is a terminal; pipe or redirect the input, for example:\n" +
		"  echo 'Hello world!' | lacre build > hello-world.html\n"
	cases := []struct {
		args           []string
		stdout, stderr string
		code           int
	}{
		{args: []string{"build"}, stderr: hint, code: 2},
		// The option error wins over the terminal (CLI-15).
		{args: []string{"build", "--bogus"}, stderr: "lacre: flag provided but not defined: -bogus\n\n" + usage, code: 2},
		// Help typed at a prompt is the commonest real call.
		{args: []string{"build", "--help"}, stdout: usage},
	}
	for _, c := range cases {
		stdin := strings.NewReader(input)
		var stdout, stderr bytes.Buffer
		code := run(c.args, stdin, &stdout, &stderr, func() bool { return true })
		if stdout.String() != c.stdout {
			t.Errorf("lacre %q: standard output is %q, want %q", c.args, stdout.Bytes(), c.stdout)
		}
		if stderr.String() != c.stderr {
			t.Errorf("lacre %q: standard error is %q, want %q", c.args, stderr.Bytes(), c.stderr)
		}
		if code != c.code {
			t.Errorf("lacre %q: the exit code is %d, want %d", c.args, code, c.code)
		}
		if stdin.Len() != len(input) {
			t.Errorf("lacre %q read standard input", c.args)
		}
	}
}

// T-13: the real program as a child process, the test binary run again in its
// "be lacre" mode (see TestMain). /dev/null is a character device, as a
// terminal is, and the check must not take it for one; a pipe is the other
// everyday input. Both cases guard the mistake of taking an input that is not
// a terminal for one. On Windows this runs GetConsoleMode on NUL and on a
// pipe.
func TestProgram(t *testing.T) {
	cases := []struct {
		name  string
		stdin io.Reader // nil gives the child os.DevNull
		input string    // what the in-process call gets
	}{
		{"os.DevNull", nil, ""},
		{"pipe", strings.NewReader("Hello world!\n"), "Hello world!\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "build")
			cmd.Env = append(os.Environ(), "LACRE_TEST_BE_LACRE=1")
			cmd.Stdin = c.stdin
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			if err := cmd.Run(); err != nil {
				t.Errorf("the program failed: %v", err)
			}
			if stderr.Len() != 0 {
				t.Errorf("standard error is %q, want nothing", stderr.Bytes())
			}
			want, _, _ := lacre(c.input, "build")
			if stdout.String() != want {
				t.Errorf("standard output is %q, want %q", stdout.Bytes(), want)
			}
		})
	}
}
