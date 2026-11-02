package app

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

var update = flag.Bool("update", false, "rewrite the golden files")

// checkGolden compares got with testdata/<name>.golden byte for byte. With
// -update it writes got there first.
func checkGolden(t *testing.T, name string, got []byte) {
	t.Helper()
	golden := filepath.Join("testdata", name+".golden")
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(golden, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("%v (run with -update to create it)", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("%s differs from %s (run with -update after intended changes)\n got: %q\nwant: %q",
			name, golden, got, want)
	}
}

// The golden pages, compared byte for byte.
func TestPageGolden(t *testing.T) {
	cases := []struct {
		name  string
		input string
		heads []string
	}{
		{"page", "<h1>Hello</h1>\n<p>Hello <b>world</b> &amp; friends!</p>\n", nil},
		{"empty", "", nil},
		{"head", "Hello world!\n", []string{`<meta name="description" content="A test page" />`, `<style>p { color: navy }</style>`}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			checkGolden(t, c.name, buildPage([]byte(c.input), "", "", c.heads))
		})
	}
}

// elements gives the element children of n, in order.
func elements(n *html.Node) []*html.Node {
	var els []*html.Node
	for c := range n.ChildNodes() {
		if c.Type == html.ElementNode {
			els = append(els, c)
		}
	}
	return els
}

// describe gives an element as its tag name and its attributes, in order,
// each value quoted so that a quote inside it cannot pass for its end.
func describe(n *html.Node) string {
	var b strings.Builder
	b.WriteString(n.Data)
	for _, a := range n.Attr {
		fmt.Fprintf(&b, " %s=%q", a.Key, a.Val)
	}
	return b.String()
}

// The page's elements, as an HTML parser reads them. A case that is not about
// the input gives none: any input would do.
func TestPageStructure(t *testing.T) {
	cases := []struct {
		name        string
		input       string
		title, lang string
		heads       []string
		wantTitle   string
		wantLang    string   // "" for no attribute on <html>
		wantHead    []string // the elements after the <style>, described
		wantBody    []string // the elements directly in <body>, described
	}{
		{name: "no options", wantTitle: "Untitled"},
		{name: "lang", lang: "ru", wantTitle: "Untitled", wantLang: "ru"},
		{name: "title with markup", title: "</title><b>x</b>&amp;", wantTitle: "</title><b>x</b>&amp;"},
		{name: "lang with a quote", lang: `en" onload="alert(1)`, wantTitle: "Untitled", wantLang: `en" onload="alert(1)`},
		{name: "two paragraphs", input: "<p>One</p><p>Two</p>", wantTitle: "Untitled", wantBody: []string{"p", "p"}},
		{name: "script", input: "<script>alert(1)</script>", wantTitle: "Untitled", wantBody: []string{"script"}},
		{
			name:      "two heads",
			heads:     []string{`<meta name="a" content="1" />`, `<meta name="b" content="2" />`},
			wantTitle: "Untitled",
			wantHead:  []string{`meta name="a" content="1"`, `meta name="b" content="2"`},
		},
		{name: "empty head", heads: []string{""}, wantTitle: "Untitled"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			doc, err := html.Parse(bytes.NewReader(buildPage([]byte(c.input), c.title, c.lang, c.heads)))
			if err != nil {
				t.Fatal(err)
			}
			if d := doc.FirstChild; d == nil || d.Type != html.DoctypeNode || d.Data != "html" {
				t.Error("the page does not start with the html doctype")
			}
			root := elements(doc)
			if len(root) != 1 {
				t.Fatalf("the document has %d elements, want 1", len(root))
			}
			wantHTML := "html"
			if c.wantLang != "" {
				wantHTML += fmt.Sprintf(" lang=%q", c.wantLang)
			}
			if got := describe(root[0]); got != wantHTML {
				t.Errorf("<html> is %s, want %s", got, wantHTML)
			}
			parts := elements(root[0])
			if len(parts) != 2 {
				t.Fatalf("<html> has %d elements, want <head> and <body>", len(parts))
			}
			head, body := parts[0], parts[1]

			var gotHead []string
			for _, el := range elements(head) {
				gotHead = append(gotHead, describe(el))
				if el.Data == "title" {
					var got string
					if el.FirstChild != nil {
						got = el.FirstChild.Data
					}
					if got != c.wantTitle {
						t.Errorf("the title is %q, want %q", got, c.wantTitle)
					}
				}
			}
			wantHead := append([]string{
				`meta charset="utf-8"`,
				`meta name="viewport" content="width=device-width, initial-scale=1"`,
				`meta name="generator" content="Lacre"`,
				"title",
				"style",
			}, c.wantHead...)
			if !slices.Equal(gotHead, wantHead) {
				t.Errorf("<head> holds %q, want %q", gotHead, wantHead)
			}

			var gotBody []string
			for _, el := range elements(body) {
				gotBody = append(gotBody, describe(el))
			}
			if !slices.Equal(gotBody, c.wantBody) {
				t.Errorf("<body> holds %q, want %q", gotBody, c.wantBody)
			}
			for n := range doc.Descendants() {
				if n.Type == html.ElementNode && n.Data == "b" {
					t.Error("the page has a <b>")
				}
			}
		})
	}
}

// Nothing in the page refers to anything outside it.
func TestPageSelfContained(t *testing.T) {
	page := buildPage([]byte("Hello world!\n"), "", "", nil)
	for _, s := range []string{"http://", "https://", "<link", "<script", "src=", "href=", "url(", "@import"} {
		if bytes.Contains(page, []byte(s)) {
			t.Errorf("the page contains %q", s)
		}
	}
}

// The title and the language escape only the five characters that
// html.EscapeString replaces: in the last row every other byte stays as it is.
// Each value, as a --head value, is never escaped.
func TestPageEscaping(t *testing.T) {
	cases := []struct{ value, title, html string }{
		{"a&b", "<title>a&amp;b</title>", `<html lang="a&amp;b">`},
		{"a<b", "<title>a&lt;b</title>", `<html lang="a&lt;b">`},
		{"a>b", "<title>a&gt;b</title>", `<html lang="a&gt;b">`},
		{`a"b`, "<title>a&#34;b</title>", `<html lang="a&#34;b">`},
		{"a'b", "<title>a&#39;b</title>", `<html lang="a&#39;b">`},
		{"not a language", "<title>not a language</title>", `<html lang="not a language">`},
		{" a\tb\x01\xff ", "<title> a\tb\x01\xff </title>", "<html lang=\" a\tb\x01\xff \">"},
	}
	for _, c := range cases {
		page := string(buildPage([]byte("Hello world!\n"), c.value, c.value, []string{c.value}))
		for _, want := range []string{"\n" + c.title + "\n", "\n" + c.html + "\n", "</style>\n" + c.value + "\n</head>\n"} {
			if !strings.Contains(page, want) {
				t.Errorf("the page for %q lacks %q", c.value, want)
			}
		}
	}
}

// Non-ASCII text comes out unchanged, in the body and in the title.
func TestPageNonASCII(t *testing.T) {
	const input = "Привет, мир! café 日本語 👋\n"
	page := string(buildPage([]byte(input), "Заметки", "", nil))
	if !strings.Contains(page, "\n<title>Заметки</title>\n") {
		t.Error("the title did not come out unchanged")
	}
	if !strings.HasSuffix(page, "\n<body>\n"+input+"</body>\n</html>\n") {
		t.Error("the input did not come out unchanged")
	}
}

// The input can be cut back out of the page. Every input of the golden,
// structure and pass-through tests, with the option values of its case where
// it has some, then a full document with a title and a --head value that look
// like the markers of the cut.
func TestPageInputBack(t *testing.T) {
	cases := []struct {
		input string
		title string
		heads []string
	}{
		{input: "<h1>Hello</h1>\n<p>Hello <b>world</b> &amp; friends!</p>\n"}, // page.golden
		{input: ""}, // empty.golden
		{ // head.golden
			input: "Hello world!\n",
			heads: []string{`<meta name="description" content="A test page" />`, `<style>p { color: navy }</style>`},
		},
		{input: "<p>One</p><p>Two</p>"},                        // two paragraphs
		{input: "<script>alert(1)</script>"},                   // script
		{input: "one\r\ntwo\r\n\r\nthree\r\n"},                 // CRLF
		{input: "Привет, мир! café 日本語 👋\n", title: "Заметки"}, // non-ASCII
		{input: "\xff\xfe\x00 ok\n"},                           // invalid UTF-8
		{input: "\ufeffHello world!\n"},                        // byte order mark
		{
			input: "<html><head></head>\n<body>\n<p>x</p></body></html>\n",
			title: "</head>",
			heads: []string{"<body>"},
		},
	}
	for _, c := range cases {
		page := buildPage([]byte(c.input), c.title, "", c.heads)
		_, rest, found := bytes.Cut(page, []byte("</head>\n<body>\n"))
		got, ended := bytes.CutSuffix(rest, []byte("</body>\n</html>\n"))
		if !found || !ended || string(got) != c.input {
			t.Errorf("the cut gives %q, want %q", got, c.input)
		}
	}
}

// An empty value counts as not given, byte for byte: TestPageStructure
// compares only the elements of the page, so it cannot see the blank line an
// empty --head value would add.
func TestPageEmptyValues(t *testing.T) {
	input := []byte("Hello world!\n")
	want := buildPage(input, "Untitled", "", nil)
	if got := buildPage(input, "", "", []string{"", ""}); !bytes.Equal(got, want) {
		t.Errorf("empty values change the page:\n got: %q\nwant: %q", got, want)
	}
}
