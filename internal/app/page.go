package app

import (
	"bytes"
	"html"
	"strings"
)

// markerBegin and markerEnd mark where the input starts and ends in the page,
// whatever the skin: the Body value is markerBegin, a line break, the input and
// markerEnd, with no line break added before it. lacre writes them and does not
// look for them: an input, a --head value or a skin can hold the same text, and
// a skin can leave the Body value out.
const (
	markerBegin = "<!-- lacre:begin -->"
	markerEnd   = "<!-- lacre:end -->"
)

// buildPage makes the page by running the skin's template with four values:
// Title, Language, Head and Body. The input and the --head values go in as
// they are, unchecked; the title and the language are escaped with
// html.EscapeString. An empty value counts as not given: the title becomes
// "Untitled", Language is empty, and an empty --head value adds nothing. Head
// is each non-empty --head value followed by a line break. The only error is the
// template's, such as a value that does not exist or a file that include or
// dataURI cannot read.
func buildPage(s *skin, input []byte, title, lang string, heads []string) ([]byte, error) {
	if title == "" {
		title = "Untitled"
	}
	var head strings.Builder
	for _, h := range heads {
		if h != "" {
			head.WriteString(h)
			head.WriteByte('\n')
		}
	}
	var page bytes.Buffer
	err := s.page.Execute(&page, map[string]string{
		"Title":    html.EscapeString(title),
		"Language": html.EscapeString(lang),
		"Head":     head.String(),
		"Body":     markerBegin + "\n" + string(input) + markerEnd,
	})
	if err != nil {
		return nil, err
	}
	return page.Bytes(), nil
}
