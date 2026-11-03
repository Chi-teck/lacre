# Usage

`lacre build` reads all of standard input and writes the page to standard output. Messages go to standard error.

```bash
lacre build --title 'My page' --lang en < snippet.html > page.html
```

The page is a complete HTML document. Its head has the UTF-8 charset, a viewport tag so that the page fits a phone's
screen, a generator tag that names lacre, the title (`Untitled` without `--title`) and a small built-in stylesheet. It
fixes what browsers draw badly by default: a centered column that is easy to read, in the system fonts at a slightly
larger size, with lists, quotations, tables, code, forms and disclosures that are easy to tell apart, long lines in
`<pre>` that scroll inside their box, and images and videos that shrink to fit the column. The page is light;
`lacre style-guide` lists what the stylesheet does and how an input can follow a dark system setting. Its body is your
input, as it is, between two markers. That is the page of the built-in theme, and a [theme](themes.md) of your own can
change all of it. With the built-in theme, nothing lacre adds refers to anything outside the page, so the page is a
single file that works offline, unless what you put in it refers to outside files. The same input, options and theme
always give the same bytes.

lacre writes `<!-- lacre:begin -->` and a line break before the input and `<!-- lacre:end -->` right after it, with no
line break added, so that the input can be found in the page. lacre does not check for the same text elsewhere: in the
input, in a `--head` value or in a theme. With `echo 'Hello world!' | lacre build`, the end of the page is:

```html
<body>
<!-- lacre:begin -->
Hello world!
<!-- lacre:end -->
</body>
</html>
```

Without `--lang` the page declares no language, since a wrong one is worse than none. Setting it helps screen readers
and satisfies accessibility checkers.

`--head` adds HTML at the end of `<head>`, as it is, and can be given more than once. The built-in stylesheet puts
its rules in a cascade layer, so a `<style>` added this way overrides them, whatever its selectors:

```bash
lacre build --head '<style>body { max-width: 60rem }</style>' < snippet.html > page.html
```

To let the page follow a dark system setting, for an input that sets no colors of its own:

```bash
lacre build --head '<style>:root { color-scheme: light dark }</style>' < snippet.html > page.html
```

To add a file, let the shell read it:

```bash
lacre build --head "$(cat head.html)" < snippet.html > page.html
```

A `--head` value that refers to outside files, such as a `<link>` to a stylesheet or a `<script src>`, makes the page
depend on them.

## Options

The options of `build`; `style-guide` takes only `--theme` and `--help`, and `install-skill` only `--help`:

| Option         | Value       | Default                      | Effect                                                                                       |
|----------------|-------------|------------------------------|----------------------------------------------------------------------------------------------|
| `--title TEXT` | any string  | `Untitled`                   | The page title                                                                               |
| `--lang CODE`  | any string  | none                         | The page language, the `lang` attribute of `<html>`; without it there is no `lang` attribute |
| `--head HTML`  | any string  | none                         | HTML added at the end of `<head>`, as it is; can be repeated                                 |
| `--theme DIR`  | a directory | `LACRE_THEME`, else built-in | The [theme](themes.md) directory; empty, as in `--theme=`, for the built-in theme            |
| `--help`, `-h` | none        |                              | Print the usage to standard output and exit with 0                                           |

Before the command, the only options are `--help` (or `-h`) and `--version`: `lacre --help` prints the usage and
`lacre --version` the version.

## The input is HTML

- The input goes into the page as it is and is not cleaned: a script in the input runs when the page is opened. Use
  lacre only on input you trust. The same goes for `--head` values.
- lacre does not repair broken HTML. A snippet with unclosed tags, or a full document with its own `<html>`, goes in as
  it is, and the browser makes of it what it can.
- Plain text is read as HTML too: its line breaks collapse into one paragraph, and `<` and `&` are read as markup. To
  show it as typed, escape it and wrap it in `<pre>`. In Unix shells:

  ```bash
  { echo '<pre>'; sed -e 's/&/\&amp;/g' -e 's/</\&lt;/g' -e 's/>/\&gt;/g' notes.txt; echo '</pre>'; } \
    | lacre build > notes.html
  ```

  With the built-in theme a long line in a `<pre>` scrolls sideways. For text with a paragraph on each line, make it
  wrap with `--head '<style>pre { white-space: pre-wrap }</style>'`.

## Style guide

`lacre style-guide` prints the style guide of the theme in use: what the input can use, such as a framework and its
classes, and what the page already puts around the input. The theme's author writes it in Markdown for whoever writes
the input, often an AI agent, and lacre prints it as it is. Read it before writing input for a theme:

```bash
lacre style-guide
```

It picks the theme as `build` does, from `--theme` or else `LACRE_THEME`, so the plain call describes the theme that
`lacre build` uses. It reads no input, so it runs at a terminal too. A theme with no style guide makes it fail with exit
code 2; the built-in theme has one. [Themes](themes.md#style-guide) tells how to write one.

## Agent skill

An AI agent uses lacre only if it knows about it; otherwise it writes the whole page by hand. `lacre install-skill DIR`
writes a skill, a file that tells the agent when to use lacre and how: read the style guide, write the content, build
the page. Agents read skills from directories of their own:

```bash
lacre install-skill .agents/skills       # for this project, for agents that read .agents/skills
lacre install-skill ~/.claude/skills     # for every project of this user, in Claude Code
```

`DIR` is the directory that holds the agent's skills, not the skill's own: the skill goes into `DIR/lacre/SKILL.md`.
lacre creates the directories it needs, so the command works in a fresh project too. It replaces an existing
`SKILL.md`, so running it again after an upgrade updates the skill, and edits made to the installed copy are lost;
other files in the `lacre` directory stay. The `version` in the skill's front matter names the lacre that wrote it, so
that a copy left behind by an older lacre can be spotted.

On success, lacre prints a line on standard error that names the file it wrote, so a mistyped path shows there:

```
lacre: wrote .agents/skills/lacre/SKILL.md
```

The shell expands `~` only when it is not quoted, and `cmd.exe` and PowerShell before 7.6 not at all. lacre does not
expand it either, so rather than create a directory named `~` in the current directory, it rejects a `DIR` whose first
name is `~`, as in `lacre install-skill '~/.claude/skills'`, with exit code 2. Name the home directory instead, as in
`lacre install-skill "$HOME/.claude/skills"`; on Windows, see [Windows](windows.md). The command reads no input, so it
runs at a terminal too.

Where common agents read skills, checked on 2026-11-03:

| Agent       | Project                            | User                                   |
|-------------|------------------------------------|----------------------------------------|
| Claude Code | `.claude/skills`                   | `~/.claude/skills`                     |
| Codex       | `.agents/skills`                   | `~/.agents/skills`                     |
| Gemini CLI  | `.agents/skills`, `.gemini/skills` | `~/.agents/skills`, `~/.gemini/skills` |

GitHub Copilot, Cursor and opencode read `.agents/skills` and `~/.agents/skills` too. Claude Code reads neither, so to
set up Claude Code and another agent, run the command once for each agent's directory.

## Exit codes

| Code | When                                                                                                                                                                                                                                                                   |
|------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0    | The page or the style guide was written, the skill was installed, or the usage or version was printed                                                                                                                                                                  |
| 1    | Reading standard input or writing standard output failed, or `install-skill` could not create a directory or write the skill                                                                                                                                           |
| 2    | The tool was called wrongly: no command, an unknown command, a bad option, a missing, empty or extra argument, a `~` that the shell did not expand, a terminal or another character device as input, a broken theme or, for `style-guide`, a theme with no style guide |

## Known limits

When standard input is a terminal, `lacre build` does not wait for typed input: it prints a hint and exits with 2. On
Linux, macOS and other Unix systems, lacre takes any character device other than `/dev/null` for a terminal, so it
refuses `/dev/zero` or `/dev/urandom` the same way. In Git Bash (mintty), lacre cannot tell that standard input is a
terminal, so it waits for input, as `cat` does. Ctrl+D ends the input. On systems that are neither Unix nor Windows,
such as Plan 9, lacre does not check at all and waits the same way.
