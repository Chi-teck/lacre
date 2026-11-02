package app

import (
	"fmt"
	"html"
	"strings"
)

// pageTemplate is the page (PAGE-1), with a %s for each of its five parts, in
// order: the lang attribute, the title, the stylesheet, the --head values and
// the body. Its generator tag names the tool and has no version (D8); its meta
// tags end with " />" (D9). It holds no comment, version or date (PAGE-9), and
// nothing that refers outside the page (PAGE-8). A Go raw string drops carriage
// returns, so the page has LF line endings on every system.
//
// The input can be cut back out of the page (PAGE-11): it is everything after
// the first "</head>\n<body>\n" and before the "</body>\n</html>\n" that end
// the page. A change here keeps that shape; TestPageInputBack checks it.
const pageTemplate = `<!DOCTYPE html>
<html%s>
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<meta name="generator" content="Lacre" />
<title>%s</title>
<style>
%s
</style>
%s</head>
<body>
%s</body>
</html>
`

// stylesheet is the page's only styling (PAGE-6, PAGE-7). Element selectors
// only, so that any style the input or a --head value brings wins; system
// fonts; a centred column; no colours and no rules for print. ui-monospace
// comes first because browsers shrink text whose only font is monospace to 13
// pixels. At 0.875em a line of 80 monospace characters fits the 48rem column,
// and pre-wrap and break-word wrap longer lines instead of cutting them off.
//
// It goes into pageTemplate as an argument, never as part of it, because of
// the % in 100%.
const stylesheet = `body {
  max-width: 48rem;
  margin: 0 auto;
  padding: 1rem;
  font-family: system-ui, sans-serif;
  line-height: 1.5;
  overflow-wrap: break-word;
}
pre {
  font-family: ui-monospace, monospace;
  font-size: 0.875em;
  white-space: pre-wrap;
}
img, video {
  max-width: 100%;
  height: auto;
}`

// buildPage makes the page from the input and the option values. The input
// and the --head values go in as they are, unchecked (PAGE-2, PAGE-3,
// PAGE-10); the title and the language are escaped with html.EscapeString
// (PAGE-5). An empty value counts as not given (CLI-4): the title becomes
// "Untitled", the page gets no lang attribute, and an empty --head value adds
// nothing.
func buildPage(input []byte, title, lang string, heads []string) []byte {
	if title == "" {
		title = "Untitled"
	}
	if lang != "" {
		lang = ` lang="` + html.EscapeString(lang) + `"`
	}
	var head strings.Builder
	for _, h := range heads {
		if h != "" {
			head.WriteString(h)
			head.WriteByte('\n')
		}
	}
	return fmt.Appendf(nil, pageTemplate, lang, html.EscapeString(title), stylesheet, head.String(), input)
}
