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

// builtinSkin is the skin in skin/, the one lacre uses when LACRE_SKIN is unset
// or empty. It is a skin like any other: openSkin reads it with the same code
// as a user skin, and it uses only the four values and the two functions that
// every skin has.
//
// Its page.html has a generator tag that names the tool and has no version,
// and meta tags that end with " />". The page holds no comment other than the
// two markers, no version or date, and nothing that refers outside the page.
// Its files have LF line endings, which .gitattributes keeps in a Windows
// checkout too, so the page has them on every system; TestBuiltinSkinNoCR
// checks it.
//
// style.css is the page's only styling, and these notes are here because a
// comment in it would end up in the page. Element selectors only, so that any
// style the input or a --head value brings wins; system fonts; a centred
// column; no colours and no rules for print. ui-monospace comes first because
// browsers shrink text whose only font is monospace to 13 pixels. At 0.875em a
// line of 80 monospace characters fits the 48rem column, and pre-wrap and
// break-word wrap longer lines instead of cutting them off.
//
//go:embed skin
var builtinSkin embed.FS

// mediaTypes gives the media type of a data: URI, by the file's extension in
// lower case. The list is lacre's own rather than the system's, which differs
// between machines, so that the same skin gives the same bytes on every one.
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

// errSkinPath is the error for a path given to include or dataURI that breaks
// the rule on paths in a skin.
var errSkinPath = errors.New(`a path in a skin is names joined by /, with no \ or : and no empty, . or .. name`)

// skin is an open skin: its files and its parsed page.html.
type skin struct {
	fsys fs.FS
	root *os.Root // nil for the built-in skin
	page *template.Template
}

// openSkin opens the skin in dir, or the built-in skin when dir is empty, and
// reads and parses its page.html. A relative dir is relative to the current
// directory. The files that include and dataURI read are read later, while the
// template runs. A user skin is read through os.Root, so that no file outside
// dir can be read, not even through a symbolic link.
//
// An error from the system comes without its *fs.PathError, whose path the
// message already gives: the caller names the skin, and openSkin names
// page.html. A parse error is the template's own.
func openSkin(dir string) (*skin, error) {
	s := &skin{}
	if dir == "" {
		fsys, err := fs.Sub(builtinSkin, "skin")
		if err != nil {
			return nil, err
		}
		s.fsys = fsys
	} else {
		root, err := os.OpenRoot(dir)
		if err != nil {
			return nil, stripPath(err)
		}
		s.fsys, s.root = root.FS(), root
	}
	text, err := fs.ReadFile(s.fsys, "page.html")
	if err != nil {
		s.Close()
		return nil, fmt.Errorf("page.html: %w", stripPath(err))
	}
	// missingkey=error makes a misspelt value, such as .title, an error that
	// names the key, where a struct would give a Go type name.
	s.page, err = template.New("page.html").Option("missingkey=error").
		Funcs(template.FuncMap{"include": s.include, "dataURI": s.dataURI}).
		Parse(string(text))
	if err != nil {
		s.Close()
		return nil, err
	}
	return s, nil
}

// Close closes the skin's directory; for the built-in skin it does nothing. An
// error is dropped: nothing was written through the directory.
func (s *skin) Close() {
	if s.root != nil {
		_ = s.root.Close()
	}
}

// include gives the file at name as it is. The file is not a template: a {{
// in it is plain text.
func (s *skin) include(name string) (string, error) {
	b, err := s.readFile(name)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// dataURI gives the file at name as a data: URI, in standard base64, with the
// media type mediaTypes has for its extension. An extension not there is an
// error, rather than a guess. The path is checked first, so a bad path gets
// the path message whatever its extension.
func (s *skin) dataURI(name string) (string, error) {
	b, err := s.readFile(name)
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

// readFile reads the file at name in the skin, for include and dataURI. It
// checks name first, on every system and in the built-in skin too: \ and : are
// ordinary characters in a filename on Linux, so without the check a skin could
// work there and fail on Windows. A symbolic link that leaves the skin fails in
// os.Root, with Go's own words. Windows reserved names, such as nul, are not
// checked and fail on Windows only.
//
// An error from the system comes without its *fs.PathError: the template
// error that carries it already shows the call and its path.
func (s *skin) readFile(name string) ([]byte, error) {
	if !fs.ValidPath(name) || name == "." || strings.ContainsAny(name, `\:`) {
		return nil, errSkinPath
	}
	b, err := fs.ReadFile(s.fsys, name)
	if err != nil {
		return nil, stripPath(err)
	}
	return b, nil
}
