# lacre

[![CI](https://github.com/Chi-teck/lacre/actions/workflows/ci.yml/badge.svg)](https://github.com/Chi-teck/lacre/actions/workflows/ci.yml)

lacre turns plain text or an HTML snippet into a full HTML page. A bit of text or a snippet is not a page: to open it in
a browser, share it or print it, someone has to wrap it in the boilerplate of a complete HTML document. lacre does that,
and nothing else. The look of the page comes from a theme: the built-in one or your own.

```bash
echo 'Hello world!' | lacre build > hello-world.html
```

Opening `hello-world.html` in a browser shows "Hello world!" on a complete page.

## Why use it

HTML is a good format for documents such as reports and notes. Any program can write it, any browser can open it,
wide tables and code can scroll instead of being cut to fit a page, and a script can make it interactive. But a complete
page needs a doctype, a head and styles, which are often longer than the content itself.

lacre is made mainly for AI agents. Asked for a report, an agent often spends more time and tokens on the look of the
page than on the report's data. With lacre, the page comes from a theme and the agent writes only the content, so pages
are cheaper and faster to make and share one look, set in one place: the theme.

## Install

There is no release yet. From the first one on, the [releases page](https://github.com/Chi-teck/lacre/releases) will
have ready-made binaries for Linux, macOS and Windows, on amd64 and arm64.

With Go 1.26 or later, install it from source:

```bash
go install github.com/Chi-teck/lacre@latest
```

## Usage

`lacre build` reads all of standard input and writes the page to standard output:

```bash
lacre build --title 'My page' --lang en < snippet.html > page.html
```

The input goes into the page as it is and is not cleaned: a script in the input runs when the page is opened. Use lacre
only on input you trust.

For a look of your own, make a [theme](docs/themes.md) and name it with `--theme`, or with the `LACRE_THEME` environment
variable. `lacre --help` lists all the commands and options.

## For AI agents

An agent uses lacre only if it knows about it; otherwise it writes the whole page by hand. `lacre install-skill DIR`
installs a skill that tells the agent when to use lacre and how:

```bash
lacre install-skill .claude/skills    # for Claude Code
lacre install-skill .agents/skills    # for Codex and other agents that read .agents/skills
```

Then ask the agent for a report or notes as an HTML page. Following the skill, the agent runs `lacre style-guide` to
learn what the theme in use offers, writes only the content and builds the page with lacre.
[Usage](docs/usage.md#agent-skill) lists where other agents read skills, and how to install the skill for all of a
user's projects.

## Documentation

- [Usage](docs/usage.md): the commands and their options, the page lacre builds, the input, the style guide, the agent
  skill and exit codes
- [Themes](docs/themes.md): using a theme, and making one with your own layout, styles and fonts and a style guide for
  whoever writes the input
- [Windows](docs/windows.md): `cmd.exe`, PowerShell and Git Bash
- [Development](docs/development.md): tasks, tests, continuous integration and releases

## License

[MIT](LICENSE)
