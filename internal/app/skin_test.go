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

// writeSkin writes a skin into a new temporary directory and gives the
// directory. files maps each path in the skin, with / between names, to its
// content.
func writeSkin(t *testing.T, files map[string]string) string {
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

// wantPage checks that lacre build with the skin in dir, input and args gave
// want, with exit code 0 and nothing on standard error.
func wantPage(t *testing.T, dir, input string, args []string, want string) {
	t.Helper()
	stdout, stderr, code := lacreSkin(dir, input, append([]string{"build"}, args...)...)
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

// wantSkinError checks that lacre build with the skin in dir failed as a
// broken skin does: exit code 2, nothing on standard output, and on standard
// error "lacre: skin <dir>: " followed by msg and sys. msg is lacre's part and
// is compared on every system. sys is the part that comes from the system,
// which Windows words differently, so the whole message is compared only on
// other systems.
func wantSkinError(t *testing.T, dir, msg, sys string) {
	t.Helper()
	stdout, stderr, code := lacreSkin(dir, "Hello world!\n", "build")
	if code != 2 {
		t.Errorf("the exit code is %d, want 2", code)
	}
	if stdout != "" {
		t.Errorf("standard output is %q, want nothing", stdout)
	}
	want := "lacre: skin " + dir + ": " + msg
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

// The four values, each on a line of its own. The body is the input between
// the markers, byte for byte, with no line break added before the end marker.
func TestSkinValues(t *testing.T) {
	dir := writeSkin(t, map[string]string{
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

// A skin's text comes out as written: the comment stays and + stays +, as in
// the title C++. Go's html/template would drop the one and write the other as
// &#43;.
func TestSkinAsWritten(t *testing.T) {
	dir := writeSkin(t, map[string]string{
		"page.gohtml": "<!-- a comment -->\n<p>1 + 1</p>\n<title>{{.Title}}</title>\n",
	})
	wantPage(t, dir, "", []string{"--title", "C++"}, "<!-- a comment -->\n<p>1 + 1</p>\n<title>C++</title>\n")
}

// include gives a file's bytes exactly, from anywhere in the skin, and a {{ in
// an included file is plain text.
func TestSkinInclude(t *testing.T) {
	const css = "p { color: navy }\r\n/* {{.Title}} */\xff"
	const js = "if (a {{ b) {}\n"
	dir := writeSkin(t, map[string]string{
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
func TestSkinDataURI(t *testing.T) {
	dir := writeSkin(t, map[string]string{
		"page.gohtml":       `{{dataURI "logo.png"}} {{dataURI "fonts/inter.woff2"}} {{dataURI "icon.PNG"}}`,
		"logo.png":          "\x89PNG\xfb\xff",
		"fonts/inter.woff2": "wOF2",
		"icon.PNG":          "abc",
	})
	wantPage(t, dir, "", nil,
		"data:image/png;base64,iVBOR/v/ data:font/woff2;base64,d09GMg== data:image/png;base64,YWJj")

	t.Run("unknown extension", func(t *testing.T) {
		dir := writeSkin(t, map[string]string{"page.gohtml": `{{dataURI "logo.tiff"}}`, "logo.tiff": "II*\x00"})
		wantSkinError(t, dir,
			`page.gohtml:1:2: executing "page.gohtml" at <dataURI "logo.tiff">: error calling dataURI: no media type for ".tiff"`, "")
	})
	t.Run("missing file", func(t *testing.T) {
		dir := writeSkin(t, map[string]string{"page.gohtml": `{{dataURI "logo.png"}}`})
		wantSkinError(t, dir, `page.gohtml:1:2: executing "page.gohtml" at <dataURI "logo.png">: error calling dataURI: `,
			"no such file or directory")
	})
	t.Run("bad path", func(t *testing.T) {
		dir := writeSkin(t, map[string]string{"page.gohtml": `{{dataURI "logo.png/"}}`, "logo.png": "\x89PNG"})
		wantSkinError(t, dir, `page.gohtml:1:2: executing "page.gohtml" at <dataURI "logo.png/">: error calling dataURI: `+
			`a path in a skin is names joined by /, with no \ or : and no empty, . or .. name`, "")
	})
}

// A path in a skin is names joined by /, and cannot leave the skin. A bad path
// fails with lacre's own message on every system, before anything is read; a
// symbolic link that leaves the skin fails in os.Root, and one that stays
// inside works.
func TestSkinPaths(t *testing.T) {
	outside := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(outside, []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	const pathMsg = `a path in a skin is names joined by /, with no \ or : and no empty, . or .. name`
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
			dir := writeSkin(t, map[string]string{
				"page.gohtml":   fmt.Sprintf("{{include %q}}", c.path),
				"css/style.css": "p {}",
			})
			wantSkinError(t, dir,
				fmt.Sprintf(`page.gohtml:1:2: executing "page.gohtml" at <include %q>: error calling include: %s`, c.path, pathMsg),
				"")
		})
	}

	t.Run("symbolic link out", func(t *testing.T) {
		dir := writeSkin(t, map[string]string{"page.gohtml": `{{include "out.css"}}`})
		if err := os.Symlink(outside, filepath.Join(dir, "out.css")); err != nil {
			t.Skip(err)
		}
		wantSkinError(t, dir, `page.gohtml:1:2: executing "page.gohtml" at <include "out.css">: error calling include: `,
			"path escapes from parent")
	})
	t.Run("symbolic link in", func(t *testing.T) {
		dir := writeSkin(t, map[string]string{"page.gohtml": `{{include "link.css"}}`, "css/style.css": "p {}"})
		if err := os.Symlink(filepath.Join("css", "style.css"), filepath.Join(dir, "link.css")); err != nil {
			t.Skip(err)
		}
		wantPage(t, dir, "", nil, "p {}")
	})
}

// A broken skin: exit code 2, nothing on standard output, and a message that
// names the skin, the file and, for a template error, the line.
func TestSkinBroken(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name     string
		dir      string
		msg, sys string // lacre's part and the system's part, as wantSkinError takes them
	}{
		{name: "missing directory", dir: filepath.Join(t.TempDir(), "missing"), sys: "no such file or directory"},
		{name: "file", dir: file, sys: "not a directory"},
		{
			name: "missing page.gohtml",
			dir:  writeSkin(t, map[string]string{"style.css": "p {}"}),
			msg:  "page.gohtml: ",
			sys:  "no such file or directory",
		},
		{
			name: "parse error",
			dir:  writeSkin(t, map[string]string{"page.gohtml": "<p>\n{{foo}}\n"}),
			msg:  `page.gohtml:2: function "foo" not defined`,
		},
		{
			name: "no such value",
			dir:  writeSkin(t, map[string]string{"page.gohtml": "{{.Nope}}"}),
			msg:  `page.gohtml:1:2: executing "page.gohtml" at <.Nope>: map has no entry for key "Nope"`,
		},
		{
			name: "value in lower case",
			dir:  writeSkin(t, map[string]string{"page.gohtml": "{{.title}}"}),
			msg:  `page.gohtml:1:2: executing "page.gohtml" at <.title>: map has no entry for key "title"`,
		},
		{
			name: "missing included file",
			dir:  writeSkin(t, map[string]string{"page.gohtml": `{{include "x.css"}}`}),
			msg:  `page.gohtml:1:2: executing "page.gohtml" at <include "x.css">: error calling include: `,
			sys:  "no such file or directory",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			wantSkinError(t, c.dir, c.msg, c.sys)
		})
	}
}

// A broken skin stops lacre before the terminal check and before it reads
// standard input. --help and --version do not open the skin.
func TestSkinOrder(t *testing.T) {
	const input = "Hello world!\n"
	missing := filepath.Join(t.TempDir(), "missing")
	for _, terminal := range []bool{false, true} {
		stdin := strings.NewReader(input)
		var stdout, stderr bytes.Buffer
		code := run([]string{"build"}, missing, stdin, &stdout, &stderr, func() bool { return terminal })
		if code != 2 {
			t.Errorf("terminal %v: the exit code is %d, want 2", terminal, code)
		}
		if want := "lacre: skin " + missing + ": "; !strings.HasPrefix(stderr.String(), want) {
			t.Errorf("terminal %v: standard error is %q, want it to start with %q", terminal, stderr.Bytes(), want)
		}
		if stdin.Len() != len(input) {
			t.Errorf("terminal %v: lacre read standard input", terminal)
		}
	}

	versionLine := "lacre " + buildVersion() + "\n"
	for _, c := range []struct {
		args []string
		want string
	}{
		{[]string{"--help"}, usage},
		{[]string{"build", "--help"}, usage},
		{[]string{"--version"}, versionLine},
	} {
		stdout, stderr, code := lacreSkin(missing, input, c.args...)
		if stdout != c.want || stderr != "" || code != 0 {
			t.Errorf("lacre %q: got %q, %q and exit code %d, want %q, nothing and 0", c.args, stdout, stderr, code, c.want)
		}
	}
}

// An empty LACRE_SKIN gives the built-in page.
func TestSkinEmpty(t *testing.T) {
	want, err := os.ReadFile(filepath.Join("testdata", "page.golden"))
	if err != nil {
		t.Fatal(err)
	}
	wantPage(t, "", "<h1>Hello</h1>\n<p>Hello <b>world</b> &amp; friends!</p>\n", nil, string(want))
}

// No file of the built-in skin holds a carriage return, so the page has LF
// line endings even in a program built from a Windows checkout. .gitattributes
// keeps them out; this catches a file that got them anyway.
func TestBuiltinSkinNoCR(t *testing.T) {
	err := fs.WalkDir(builtinSkin, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := fs.ReadFile(builtinSkin, name)
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
