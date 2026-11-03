// Command lacre turns plain text or an HTML snippet into a full HTML page.
//
//	echo 'Hello world!' | lacre build > hello-world.html
//
// The build command reads all of standard input and writes the page to
// standard output. Messages go to standard error.
//
// The style-guide command prints the style guide of the theme in use, its
// style-guide.md, as it is, in Markdown. In it the theme's author tells
// whoever writes the input, often an AI agent, what the theme offers and what
// the page already puts around the input. A theme without one makes the
// command fail.
//
// The install-skill command writes a skill for AI agents, such as Claude Code
// or Codex, into the skills directory it is given, as lacre/SKILL.md in it.
// The skill tells an agent when to use lacre and how: read the style guide,
// write the content, build the page. The command creates the directories it
// needs and replaces an existing copy, so running it again after an upgrade
// updates the skill; edits made to the copy are lost. A ~ that the shell did
// not expand, as in a quoted '~/.claude/skills', makes the command fail.
//
// With the built-in theme, the page is self-contained: nothing lacre adds to it
// refers to anything outside the page. It is a single file that works offline,
// unless what you put in it refers to outside files.
//
// The --theme option, of build and style-guide, names a theme: a directory
// with a page.gohtml, the page template in Go's text/template syntax, the
// files that template brings into the page and, optionally, a style-guide.md.
// A relative path is relative to the current directory. Without --theme, the
// environment variable LACRE_THEME names it. An empty --theme, or no --theme
// and an unset or empty LACRE_THEME, gives the built-in theme.
//
// The input goes into the page as HTML, as it is, and is not cleaned: a script
// in the input runs when the page is opened, so use lacre only on input you
// trust. Plain text needs escaping, and a <pre> to keep its line breaks,
// before it goes in.
package main

import "github.com/Chi-teck/lacre/internal/app"

func main() { app.Main() }
