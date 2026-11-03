package app

import (
	"embed"
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"text/template"
)

// builtinTheme is the theme in theme/, the one lacre uses with an empty
// --theme, or without --theme when LACRE_THEME is unset or empty. It is a
// theme like any other: openTheme reads it with the same code as a user theme,
// and it uses only the four values and the two functions that every theme has.
//
// Its page.gohtml has a generator tag that names the tool and has no version,
// and meta tags that end with "/>". The include of style.css and .Head end in
// -}}, which drops the template's line break after them, since the file and
// each --head value end with their own. The page holds no comment other than
// the two markers, no version or date, and nothing that refers outside the
// page. Its files have LF line endings, which .gitattributes keeps in a
// Windows checkout too, so the page has them on every system;
// TestBuiltinThemeNoCR checks it.
//
// style.css is the page's only styling, and these notes are here because a
// comment in it would end up in the page. It fixes what browsers draw badly by
// default, such as deep list indents, tables without borders and form controls
// in a smaller font, and adds little look of its own.
//
// Every rule is in the layer lacre, so that any style the input or a --head
// value brings wins whatever its selectors, a framework's own layers included,
// since they come later. That lets the rules use compound selectors. None may
// be !important, which a layer turns around.
//
// Presentational attributes in the input, such as align on a cell or width on
// a table, lose to every rule, a layered one too, so a rule that would hide one
// skips the elements that have it. vertical-align is set on the row groups,
// and the browser's own rules pass it down to the rows and cells, so that
// valign on a row works too. th inherits text-align instead of being centered,
// so that it lines up with the cells below it and follows align on its row.
//
// The font size is a percentage, so that the page follows the reader's default
// size, and other lengths are in em or rem, so that they grow with it; only
// thin lines are 1px. Properties that name a side are logical, such as
// padding-inline-start, so that a right-to-left page mirrors. Borders and the
// backgrounds of code are a translucent gray, which shows on a light page, a
// dark one and the input's own colors, so their variables, --lacre-border,
// --lacre-pre-bg and --lacre-code-bg, need no dark twins. The text and page
// background colors are the other --lacre-* variables, picked by light-dark()
// from the scheme the page is drawn in: light, unless the input or a --head
// value declares color-scheme. The background is on body, not html: the
// browser paints the root's background over the whole window, so on html it
// would confine an input's own body background to the column.
//
// Monospace fonts come as a list because browsers shrink text whose only font
// is monospace to 13 pixels. code, kbd and samp get a background only outside
// a pre and outside one another, so that a key combination or a highlighted
// line is not a box in a box. Tables fill the column, and code and links in
// their cells break anywhere, so that a wide table does not widen the page
// while ordinary words still wrap at spaces. pre scrolls rather than wraps,
// which would mangle diagrams and indentation, and wraps in print, where a
// scroll box would be cut off. Form controls take only the font family and
// size, so that a control inside a bold label keeps its own weight, and the
// labels of checkboxes and radio buttons keep the weight of the text around
// them, since a row of bold choices reads heavy. svg gets no height: auto,
// which would turn a chart drawn at a fixed height and full width into a
// square.
//
// style-guide.md is its style guide, which lacre style-guide prints. It
// describes what page.gohtml and style.css give the input, so a change there
// that an input's author would notice changes the guide too.
//
//go:embed theme
var builtinTheme embed.FS

// mediaTypes gives the media type of a data: URI, by the file's extension in
// lower case. The list is lacre's own rather than the system's, which differs
// between machines, so that the same theme gives the same bytes on every one.
// .ico has two types in common use: image/vnd.microsoft.icon is the registered
// one, and browsers accept both.
var mediaTypes = map[string]string{
	".png":   "image/png",
	".jpg":   "image/jpeg",
	".jpeg":  "image/jpeg",
	".gif":   "image/gif",
	".webp":  "image/webp",
	".avif":  "image/avif",
	".svg":   "image/svg+xml",
	".ico":   "image/vnd.microsoft.icon",
	".bmp":   "image/bmp",
	".woff":  "font/woff",
	".woff2": "font/woff2",
	".ttf":   "font/ttf",
	".otf":   "font/otf",
}

// errThemePath is the error for a path given to include or dataURI that breaks
// the rule on paths in a theme.
var errThemePath = errors.New(`a path in a theme is names joined by /, with no \ or : and no empty, . or .. name`)

// errNoStyleGuide is what styleGuide gives for a theme without a
// style-guide.md.
var errNoStyleGuide = errors.New("no style guide: style-guide.md does not exist")

// theme is an open theme: its files and its parsed page.gohtml.
type theme struct {
	fsys fs.FS
	root *os.Root // nil for the built-in theme
	page *template.Template
}

// openTheme opens the theme in dir, or the built-in theme when dir is empty,
// and reads and parses its page.gohtml. A relative dir is relative to the
// current directory. The files that include and dataURI read are read later,
// while the template runs. A user theme is read through os.Root, so that no
// file outside dir can be read, not even through a symbolic link.
//
// An error from the system comes without its *fs.PathError, whose path the
// message already gives: the caller names the theme, and openTheme names
// page.gohtml. A parse error is the template's own.
func openTheme(dir string) (*theme, error) {
	th := &theme{}
	if dir == "" {
		fsys, err := fs.Sub(builtinTheme, "theme")
		if err != nil {
			return nil, err
		}
		th.fsys = fsys
	} else {
		root, err := os.OpenRoot(dir)
		if err != nil {
			return nil, stripPath(err)
		}
		th.fsys, th.root = root.FS(), root
	}
	text, err := fs.ReadFile(th.fsys, "page.gohtml")
	if err != nil {
		th.Close()
		return nil, fmt.Errorf("page.gohtml: %w", stripPath(err))
	}
	// missingkey=error makes a misspelt value, such as .title, an error that
	// names the key, where a struct would give a Go type name.
	th.page, err = template.New("page.gohtml").Option("missingkey=error").
		Funcs(template.FuncMap{"include": th.include, "dataURI": th.dataURI}).
		Parse(string(text))
	if err != nil {
		th.Close()
		return nil, err
	}
	return th, nil
}

// Close closes the theme's directory; for the built-in theme it does nothing.
// An error is dropped: nothing was written through the directory.
func (th *theme) Close() {
	if th.root != nil {
		_ = th.root.Close()
	}
}

// styleGuide gives the theme's style-guide.md as it is. The file is not a
// template: a {{ in it is plain text. It is read like page.gohtml, so os.Root
// keeps a symbolic link from leading out of a user theme. Only a file that does
// not exist gives errNoStyleGuide; any other error names style-guide.md and
// comes without its *fs.PathError.
func (th *theme) styleGuide() ([]byte, error) {
	b, err := fs.ReadFile(th.fsys, "style-guide.md")
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil, errNoStyleGuide
	case err != nil:
		return nil, fmt.Errorf("style-guide.md: %w", stripPath(err))
	}
	return b, nil
}

// include gives the file at name as it is. The file is not a template: a {{
// in it is plain text.
func (th *theme) include(name string) (string, error) {
	b, err := th.readFile(name)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// dataURI gives the file at name as a data: URI, in standard base64, with the
// media type mediaTypes has for its extension. An extension not there is an
// error, rather than a guess. The path is checked first, so a bad path gets
// the path message whatever its extension.
func (th *theme) dataURI(name string) (string, error) {
	b, err := th.readFile(name)
	if err != nil {
		return "", err
	}
	ext := path.Ext(name)
	mediaType, ok := mediaTypes[strings.ToLower(ext)]
	if !ok {
		return "", fmt.Errorf("no media type for %q", ext)
	}
	return "data:" + mediaType + ";base64," + base64.StdEncoding.EncodeToString(b), nil
}

// readFile reads the file at name in the theme, for include and dataURI. It
// checks name first, on every system and in the built-in theme too: \ and : are
// ordinary characters in a filename on Linux, so without the check a theme
// could work there and fail on Windows. A symbolic link that leaves the theme
// fails in os.Root, with Go's own words. Windows reserved names, such as nul,
// are not checked and fail on Windows only.
//
// An error from the system comes without its *fs.PathError: the template
// error that carries it already shows the call and its path.
func (th *theme) readFile(name string) ([]byte, error) {
	if !fs.ValidPath(name) || name == "." || strings.ContainsAny(name, `\:`) {
		return nil, errThemePath
	}
	b, err := fs.ReadFile(th.fsys, name)
	if err != nil {
		return nil, stripPath(err)
	}
	return b, nil
}
