package app

import (
	_ "embed"
	"strings"
)

// skillFile is the agent skill that install-skill writes, from skill/SKILL.md.
// It is short, about one screen, since an agent reads all of it each time it
// picks lacre. It gives the reasons to let lacre build the page, since an agent
// that knows them is less likely to fall back on writing the page by hand. It
// describes no theme: the theme changes with --theme and LACRE_THEME, so lacre
// style-guide is the only source on it. Nor does it list the options, which
// --help gives and a list here would drift from. What it says holds for every
// theme, such as how plain text goes in.
//
// Its front matter follows the Agent Skills specification, so that every agent
// that reads skills can read it: name matches the lacre directory that
// install-skill writes it to, and description, which an agent reads before it
// opens the skill, says when to use lacre. The description is folded and
// indented, since Claude Code takes the first heading instead of a description
// whose lines are not indented. Claude Code also strips the front matter
// before its model reads the skill, so the body repeats what the agent must act
// on. The file has LF line endings, which .gitattributes keeps in a Windows
// checkout too; TestSkillNoCR checks it.
//
//go:embed skill/SKILL.md
var skillFile string

// skillText gives the skill with this lacre's version in place of {{version}},
// so that a copy left behind by an older lacre can be spotted. The placeholder
// is in double quotes in the front matter, so a version of digits only, such
// as a commit hash, stays a string; no version holds a " or a \.
func skillText() string {
	return strings.ReplaceAll(skillFile, "{{version}}", buildVersion())
}
