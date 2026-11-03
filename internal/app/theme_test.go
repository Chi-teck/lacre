package app

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// writeTheme writes a theme into a new temporary directory and gives the
// directory. files maps each path in the theme, with / between names, to its
// content.
func writeTheme(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// wantPage checks that lacre build with the theme in dir, input and args gave
// want, with exit code 0 and nothing on standard error.
func wantPage(t *testing.T, dir, input string, args []string, want string) {
	t.Helper()
	stdout, stderr, code := lacreTheme(dir, input, append([]string{"build"}, args...)...)
	if code != 0 {
		t.Errorf("the exit code is %d, want 0", code)
	}
	if stderr != "" {
		t.Errorf("standard error is %q, want nothing", stderr)
	}
	if stdout != want {
		t.Errorf("standard output is %q, want %q", stdout, want)
	}
}

// wantThemeError checks that lacre build with the theme in dir failed as a
// broken theme does: exit code 2, nothing on standard output, and on standard
// error "lacre: theme <dir>: " followed by msg and sys. msg is lacre's part and
// is compared on every system. sys is the part that comes from the system,
// which Windows words differently, so the whole message is compared only on
// other systems.
func wantThemeError(t *testing.T, dir, msg, sys string) {
	t.Helper()
	wantThemeErrorArgs(t, []string{"build"}, dir, msg, sys)
}

// wantThemeErrorArgs is wantThemeError with args as the command line instead
// of build.
func wantThemeErrorArgs(t *testing.T, args []string, dir, msg, sys string) {
	t.Helper()
	stdout, stderr, code := lacreTheme(dir, "Hello world!\n", args...)
	if code != 2 {
		t.Errorf("the exit code is %d, want 2", code)
	}
	if stdout != "" {
		t.Errorf("standard output is %q, want nothing", stdout)
	}
	want := "lacre: theme " + dir + ": " + msg
	if sys != "" && runtime.GOOS == "windows" {
		if !strings.HasPrefix(stderr, want) {
			t.Errorf("standard error is %q, want it to start with %q", stderr, want)
		}
		return
	}
	if want += sys + "\n"; stderr != want {
		t.Errorf("standard error is %q, want %q", stderr, want)
	}
}

// wantStyleGuide checks that lacre style-guide with the theme in dir and args
// gave want, with exit code 0 and nothing on standard error.
func wantStyleGuide(t *testing.T, dir string, args []string, want string) {
	t.Helper()
	stdout, stderr, code := lacreTheme(dir, "", append([]string{"style-guide"}, args...)...)
	if code != 0 {
		t.Errorf("the exit code is %d, want 0", code)
	}
	if stderr != "" {
		t.Errorf("standard error is %q, want nothing", stderr)
	}
	if stdout != want {
		t.Errorf("standard output is %q, want %q", stdout, want)
	}
}

// builtinStyleGuide gives the built-in theme's style-guide.md.
func builtinStyleGuide(t *testing.T) string {
	t.Helper()
	b, err := fs.ReadFile(builtinTheme, "theme/style-guide.md")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// The four values, each on a line of its own. The body is the input between
// the markers, byte for byte, with no line break added before the end marker.
func TestThemeValues(t *testing.T) {
	dir := writeTheme(t, map[string]string{
		"page.gohtml": "Title={{.Title}}\nLanguage={{.Language}}\nHead={{.Head}}\nBody={{.Body}}\n",
	})
	cases := []struct {
		name  string
		args  []string
		input string
		want  string
	}{
		{
			name:  "no options",
			input: "Hello world!\n",
			want:  "Title=Untitled\nLanguage=\nHead=\nBody=<!-- lacre:begin -->\nHello world!\n<!-- lacre:end -->\n",
		},
		{
			name:  "every option",
			args:  []string{"--title", "A & B", "--lang", "en", "--head", "<a>", "--head", "", "--head", "<b>"},
			input: "one\r\ntwo",
			want:  "Title=A &amp; B\nLanguage=en\nHead=<a>\n<b>\n\nBody=<!-- lacre:begin -->\none\r\ntwo<!-- lacre:end -->\n",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			wantPage(t, dir, c.input, c.args, c.want)
		})
	}
}

// A theme's text comes out as written: the comment stays and + stays +, as in
// the title C++. Go's html/template would drop the one and write the other as
// &#43;.
func TestThemeAsWritten(t *testing.T) {
	dir := writeTheme(t, map[string]string{
		"page.gohtml": "<!-- a comment -->\n<p>1 + 1</p>\n<title>{{.Title}}</title>\n",
	})
	wantPage(t, dir, "", []string{"--title", "C++"}, "<!-- a comment -->\n<p>1 + 1</p>\n<title>C++</title>\n")
}

// include gives a file's bytes exactly, from anywhere in the theme, and a {{ in
// an included file is plain text.
func TestThemeInclude(t *testing.T) {
	const css = "p { color: navy }\r\n/* {{.Title}} */\xff"
	const js = "if (a {{ b) {}\n"
	dir := writeTheme(t, map[string]string{
		"page.gohtml": `[{{include "style.css"}}][{{include "js/app.js"}}]`,
		"style.css":   css,
		"js/app.js":   js,
	})
	wantPage(t, dir, "", nil, "["+css+"]["+js+"]")
}

// dataURI gives the media type of the extension, in any case, and the bytes in
// standard base64, padding included. An extension lacre has no type for and a
// missing file are errors, and a bad path gets the path message even when it
// has no extension.
func TestThemeDataURI(t *testing.T) {
	dir := writeTheme(t, map[string]string{
		"page.gohtml":       `{{dataURI "logo.png"}} {{dataURI "fonts/inter.woff2"}} {{dataURI "icon.PNG"}}`,
		"logo.png":          "\x89PNG\xfb\xff",
		"fonts/inter.woff2": "wOF2",
		"icon.PNG":          "abc",
	})
	wantPage(t, dir, "", nil,
		"data:image/png;base64,iVBOR/v/ data:font/woff2;base64,d09GMg== data:image/png;base64,YWJj")

	t.Run("unknown extension", func(t *testing.T) {
		dir := writeTheme(t, map[string]string{"page.gohtml": `{{dataURI "logo.tiff"}}`, "logo.tiff": "II*\x00"})
		wantThemeError(t, dir,
			`page.gohtml:1:2: executing "page.gohtml" at <dataURI "logo.tiff">: error calling dataURI: no media type for ".tiff"`, "")
	})
	t.Run("missing file", func(t *testing.T) {
		dir := writeTheme(t, map[string]string{"page.gohtml": `{{dataURI "logo.png"}}`})
		wantThemeError(t, dir, `page.gohtml:1:2: executing "page.gohtml" at <dataURI "logo.png">: error calling dataURI: `,
			"no such file or directory")
	})
	t.Run("bad path", func(t *testing.T) {
		dir := writeTheme(t, map[string]string{"page.gohtml": `{{dataURI "logo.png/"}}`, "logo.png": "\x89PNG"})
		wantThemeError(t, dir, `page.gohtml:1:2: executing "page.gohtml" at <dataURI "logo.png/">: error calling dataURI: `+
			`a path in a theme is names joined by /, with no \ or : and no empty, . or .. name`, "")
	})
}

// A path in a theme is names joined by /, and cannot leave the theme. A bad
// path fails with lacre's own message on every system, before anything is read;
// a symbolic link that leaves the theme fails in os.Root, and one that stays
// inside works.
func TestThemePaths(t *testing.T) {
	outside := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(outside, []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	const pathMsg = `a path in a theme is names joined by /, with no \ or : and no empty, . or .. name`
	for _, c := range []struct{ name, path string }{
		{"up", "../secret.txt"},
		{"only up", ".."},
		{"only dot", "."},
		{"empty", ""},
		{"absolute", outside},
		{"dot", "./css/style.css"},
		{"double slash", "css//style.css"},
		{"trailing slash", "css/"},
		{"backslash", `css\style.css`},
		{"colon", "css:style.css"},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir := writeTheme(t, map[string]string{
				"page.gohtml":   fmt.Sprintf("{{include %q}}", c.path),
				"css/style.css": "p {}",
			})
			wantThemeError(t, dir,
				fmt.Sprintf(`page.gohtml:1:2: executing "page.gohtml" at <include %q>: error calling include: %s`, c.path, pathMsg),
				"")
		})
	}

	t.Run("symbolic link out", func(t *testing.T) {
		dir := writeTheme(t, map[string]string{"page.gohtml": `{{include "out.css"}}`})
		if err := os.Symlink(outside, filepath.Join(dir, "out.css")); err != nil {
			t.Skip(err)
		}
		wantThemeError(t, dir, `page.gohtml:1:2: executing "page.gohtml" at <include "out.css">: error calling include: `,
			"path escapes from parent")
	})
	t.Run("symbolic link in", func(t *testing.T) {
		dir := writeTheme(t, map[string]string{"page.gohtml": `{{include "link.css"}}`, "css/style.css": "p {}"})
		if err := os.Symlink(filepath.Join("css", "style.css"), filepath.Join(dir, "link.css")); err != nil {
			t.Skip(err)
		}
		wantPage(t, dir, "", nil, "p {}")
	})
}

// A broken theme: exit code 2, nothing on standard output, and a message that
// names the theme, the file and, for a template error, the line.
func TestThemeBroken(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name     string
		dir      string
		msg, sys string // lacre's part and the system's part, as wantThemeError takes them
	}{
		{name: "missing directory", dir: filepath.Join(t.TempDir(), "missing"), sys: "no such file or directory"},
		{name: "file", dir: file, sys: "not a directory"},
		{
			name: "missing page.gohtml",
			dir:  writeTheme(t, map[string]string{"style.css": "p {}"}),
			msg:  "page.gohtml: ",
			sys:  "no such file or directory",
		},
		{
			name: "parse error",
			dir:  writeTheme(t, map[string]string{"page.gohtml": "<p>\n{{foo}}\n"}),
			msg:  `page.gohtml:2: function "foo" not defined`,
		},
		{
			name: "no such value",
			dir:  writeTheme(t, map[string]string{"page.gohtml": "{{.Nope}}"}),
			msg:  `page.gohtml:1:2: executing "page.gohtml" at <.Nope>: map has no entry for key "Nope"`,
		},
		{
			name: "value in lower case",
			dir:  writeTheme(t, map[string]string{"page.gohtml": "{{.title}}"}),
			msg:  `page.gohtml:1:2: executing "page.gohtml" at <.title>: map has no entry for key "title"`,
		},
		{
			name: "missing included file",
			dir:  writeTheme(t, map[string]string{"page.gohtml": `{{include "x.css"}}`}),
			msg:  `page.gohtml:1:2: executing "page.gohtml" at <include "x.css">: error calling include: `,
			sys:  "no such file or directory",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			wantThemeError(t, c.dir, c.msg, c.sys)
		})
	}
}

// A broken theme stops either command with one line and no usage, before the
// terminal check and before build reads standard input; style-guide does
// neither. With --theme, it is that directory that is opened, not
// LACRE_THEME's. A mistake elsewhere on the command line is reported before any
// theme is opened, and --help and --version do not open one.
func TestThemeOrder(t *testing.T) {
	const input = "Hello world!\n"
	missing := filepath.Join(t.TempDir(), "missing")
	other := writeTheme(t, map[string]string{"page.gohtml": "{{.Body}}"})
	for _, c := range []struct {
		themeDir string
		args     []string
	}{
		{missing, []string{"build"}},
		{other, []string{"build", "--theme", missing}},
		{missing, []string{"style-guide"}},
		{other, []string{"style-guide", "--theme", missing}},
	} {
		for _, terminal := range []bool{false, true} {
			stdin := strings.NewReader(input)
			var stdout, stderr bytes.Buffer
			code := run(c.args, c.themeDir, stdin, &stdout, &stderr, func() bool { return terminal })
			if code != 2 {
				t.Errorf("lacre %q, terminal %v: the exit code is %d, want 2", c.args, terminal, code)
			}
			if stdout.Len() != 0 {
				t.Errorf("lacre %q, terminal %v: standard output is %q, want nothing", c.args, terminal, stdout.Bytes())
			}
			want := "lacre: theme " + missing + ": "
			if s := stderr.String(); !strings.HasPrefix(s, want) || strings.Index(s, "\n") != len(s)-1 {
				t.Errorf("lacre %q, terminal %v: standard error is %q, want one line that starts with %q",
					c.args, terminal, s, want)
			}
			if stdin.Len() != len(input) {
				t.Errorf("lacre %q, terminal %v: lacre read standard input", c.args, terminal)
			}
		}
	}

	usageErr := func(msg string) string { return "lacre: " + msg + "\n\n" + usage }
	versionLine := "lacre " + buildVersion() + "\n"
	for _, c := range []struct {
		args           []string
		stdout, stderr string
		code           int
	}{
		{args: []string{"--help"}, stdout: usage},
		{args: []string{"build", "--help"}, stdout: usage},
		{args: []string{"--version"}, stdout: versionLine},
		{args: []string{"build", "--theme", missing, "--help"}, stdout: usage},
		{
			args:   []string{"build", "--theme", missing, "--bogus"},
			stderr: usageErr("flag provided but not defined: -bogus"),
			code:   2,
		},
		{
			args:   []string{"build", "--theme", missing, "page.html"},
			stderr: usageErr("build takes no arguments; it reads standard input"),
			code:   2,
		},
		{args: []string{"style-guide", "--theme", missing, "--help"}, stdout: usage},
		{
			args:   []string{"style-guide", "--theme", missing, "--bogus"},
			stderr: usageErr("flag provided but not defined: -bogus"),
			code:   2,
		},
		{
			args:   []string{"style-guide", "--theme", missing, "extra"},
			stderr: usageErr("style-guide takes no arguments; name the theme with --theme"),
			code:   2,
		},
	} {
		stdout, stderr, code := lacreTheme(missing, input, c.args...)
		if stdout != c.stdout || stderr != c.stderr || code != c.code {
			t.Errorf("lacre %q: got %q, %q and exit code %d, want %q, %q and %d",
				c.args, stdout, stderr, code, c.stdout, c.stderr, c.code)
		}
	}
}

// --theme names the theme and wins over LACRE_THEME, whose directory is then
// not opened, so a missing one does no harm. An empty --theme, given with = or
// as an argument of its own, gives the built-in page, and the last --theme
// wins.
func TestThemeOption(t *testing.T) {
	const input = "Hello world!\n"
	a := writeTheme(t, map[string]string{"page.gohtml": "A {{.Body}}"})
	b := writeTheme(t, map[string]string{"page.gohtml": "B {{.Body}}"})
	missing := filepath.Join(t.TempDir(), "missing")
	pageB := "B <!-- lacre:begin -->\n" + input + "<!-- lacre:end -->"
	builtin := string(builtinPage(t, []byte(input), "", "", nil))
	cases := []struct {
		name     string
		themeDir string // the value of LACRE_THEME
		args     []string
		want     string
	}{
		{"over LACRE_THEME", a, []string{"--theme", b}, pageB},
		{"empty LACRE_THEME", "", []string{"--theme", b}, pageB},
		{"empty with =", a, []string{"--theme="}, builtin},
		{"empty argument", a, []string{"--theme", ""}, builtin},
		{"last one empty", a, []string{"--theme", b, "--theme="}, builtin},
		{"last one a theme", a, []string{"--theme=", "--theme", b}, pageB},
		{"missing LACRE_THEME", missing, []string{"--theme", b}, pageB},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			wantPage(t, c.themeDir, input, c.args, c.want)
		})
	}
}

// The style guide comes out as written, byte for byte: a carriage return,
// invalid UTF-8 and no final line break stay, and a {{ in it is plain text.
// The template is not run, or its include of a missing file would fail. A
// symbolic link that stays inside the theme works.
func TestStyleGuide(t *testing.T) {
	const guide = "# Guide\r\n\n{{.Title}} & <b>\n\xff"
	dir := writeTheme(t, map[string]string{
		"page.gohtml":    `{{include "missing.css"}}`,
		"style-guide.md": guide,
	})
	wantStyleGuide(t, dir, nil, guide)

	t.Run("symbolic link in", func(t *testing.T) {
		dir := writeTheme(t, map[string]string{"page.gohtml": "{{.Body}}", "docs/guide.md": guide})
		if err := os.Symlink(filepath.Join("docs", "guide.md"), filepath.Join(dir, "style-guide.md")); err != nil {
			t.Skip(err)
		}
		wantStyleGuide(t, dir, nil, guide)
	})
}

// style-guide picks the theme as build does: --theme wins over LACRE_THEME,
// whose directory is then not opened, an empty --theme gives the built-in
// theme, and the last --theme wins. TestThemeOption covers the rest, through
// the same themeOption.
func TestStyleGuideTheme(t *testing.T) {
	const guideA, guideB = "Guide A\n", "Guide B\n"
	a := writeTheme(t, map[string]string{"page.gohtml": "A {{.Body}}", "style-guide.md": guideA})
	b := writeTheme(t, map[string]string{"page.gohtml": "B {{.Body}}", "style-guide.md": guideB})
	missing := filepath.Join(t.TempDir(), "missing")
	builtin := builtinStyleGuide(t)
	cases := []struct {
		name     string
		themeDir string // the value of LACRE_THEME
		args     []string
		want     string
	}{
		{"LACRE_THEME", a, nil, guideA},
		{"over LACRE_THEME", a, []string{"--theme", b}, guideB},
		{"empty with =", a, []string{"--theme="}, builtin},
		{"last one empty", a, []string{"--theme", b, "--theme="}, builtin},
		{"last one a theme", a, []string{"--theme=", "--theme", b}, guideB},
		{"missing LACRE_THEME", missing, []string{"--theme", b}, guideB},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			wantStyleGuide(t, c.themeDir, c.args, c.want)
		})
	}
}

// Only a style-guide.md that does not exist is no style guide, and that is one
// line that says so. Any other error in reading it fails as a broken theme
// does and names style-guide.md, so that a symbolic link out of the theme
// never counts as no style guide, even when its target does not exist. The
// theme is opened first, so a broken page.gohtml fails as it does for build.
func TestStyleGuideErrors(t *testing.T) {
	outside := filepath.Join(t.TempDir(), "guide.md")
	if err := os.WriteFile(outside, []byte("# Secret\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	args := []string{"style-guide"}
	const noGuide = "no style guide: style-guide.md does not exist"

	t.Run("no style-guide.md", func(t *testing.T) {
		dir := writeTheme(t, map[string]string{"page.gohtml": "{{.Body}}"})
		wantThemeErrorArgs(t, args, dir, noGuide, "")
	})
	t.Run("dangling symbolic link in", func(t *testing.T) {
		dir := writeTheme(t, map[string]string{"page.gohtml": "{{.Body}}"})
		if err := os.Symlink("missing.md", filepath.Join(dir, "style-guide.md")); err != nil {
			t.Skip(err)
		}
		wantThemeErrorArgs(t, args, dir, noGuide, "")
	})
	for _, c := range []struct{ name, target string }{
		{"symbolic link out", outside},
		{"symbolic link out to a missing file", filepath.Join(t.TempDir(), "missing.md")},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir := writeTheme(t, map[string]string{"page.gohtml": "{{.Body}}"})
			if err := os.Symlink(c.target, filepath.Join(dir, "style-guide.md")); err != nil {
				t.Skip(err)
			}
			wantThemeErrorArgs(t, args, dir, "style-guide.md: ", "path escapes from parent")
		})
	}
	t.Run("directory", func(t *testing.T) {
		dir := writeTheme(t, map[string]string{"page.gohtml": "{{.Body}}", "style-guide.md/guide.md": "# Guide\n"})
		wantThemeErrorArgs(t, args, dir, "style-guide.md: ", "is a directory")
	})
	t.Run("unreadable", func(t *testing.T) {
		dir := writeTheme(t, map[string]string{"page.gohtml": "{{.Body}}", "style-guide.md": "# Guide\n"})
		guide := filepath.Join(dir, "style-guide.md")
		if err := os.Chmod(guide, 0); err != nil {
			t.Fatal(err)
		}
		if _, err := os.ReadFile(guide); err == nil {
			t.Skip("the file can still be read, as on Windows or as root")
		}
		wantThemeErrorArgs(t, args, dir, "style-guide.md: ", "permission denied")
	})
	t.Run("broken page.gohtml", func(t *testing.T) {
		dir := writeTheme(t, map[string]string{"page.gohtml": "<p>\n{{foo}}\n"})
		wantThemeErrorArgs(t, args, dir, `page.gohtml:2: function "foo" not defined`, "")
	})
	t.Run("same as build", func(t *testing.T) {
		for _, dir := range []string{
			filepath.Join(t.TempDir(), "missing"),
			writeTheme(t, map[string]string{"style.css": "p {}"}),
			writeTheme(t, map[string]string{"page.gohtml": "<p>\n{{foo}}\n"}),
		} {
			_, buildErr, buildCode := lacreTheme(dir, "Hello world!\n", "build")
			_, guideErr, guideCode := lacreTheme(dir, "Hello world!\n", "style-guide")
			if guideErr != buildErr || guideCode != buildCode {
				t.Errorf("theme %s: style-guide gave %q and exit code %d, build %q and %d",
					dir, guideErr, guideCode, buildErr, buildCode)
			}
		}
	})
}

// An empty LACRE_THEME gives the built-in page.
func TestThemeEmpty(t *testing.T) {
	want, err := os.ReadFile(filepath.Join("testdata", "page.golden"))
	if err != nil {
		t.Fatal(err)
	}
	wantPage(t, "", "<h1>Hello</h1>\n<p>Hello <b>world</b> &amp; friends!</p>\n", nil, string(want))
}

// No file of the built-in theme holds a carriage return, so the page has LF
// line endings even in a program built from a Windows checkout. .gitattributes
// keeps them out; this catches a file that got them anyway.
func TestBuiltinThemeNoCR(t *testing.T) {
	err := fs.WalkDir(builtinTheme, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := fs.ReadFile(builtinTheme, name)
		if err != nil {
			return err
		}
		if bytes.ContainsRune(b, '\r') {
			t.Errorf("%s holds a carriage return", name)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
