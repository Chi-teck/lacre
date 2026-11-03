# lacre

[![CI](https://github.com/Chi-teck/lacre/actions/workflows/ci.yml/badge.svg)](https://github.com/Chi-teck/lacre/actions/workflows/ci.yml)

## What it's for

lacre turns plain text or an HTML snippet into a full HTML page. A bit of text or a snippet is not a page: to open it in
a browser, share it or print it, someone has to wrap it in the boilerplate of a complete HTML document. lacre does that,
and nothing else.

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
are cheaper and faster to make, share one look, and all change when you change the theme.

## Install

Ready-made binaries for Linux, macOS and Windows, on amd64 and arm64, will be on the
[releases page](https://github.com/Chi-teck/lacre/releases) from the first release on. There is no release yet.

With Go 1.26 or later:

```bash
go install github.com/Chi-teck/lacre@latest
```

## Usage

`lacre build` reads all of standard input and writes the page to standard output. `lacre build --help` lists the
options.

```bash
lacre build --title 'My page' --lang en < snippet.html > page.html
```

The input goes into the page as it is and is not cleaned: a script in the input runs when the page is opened. Use lacre
only on input you trust.

`lacre style-guide` prints what the theme in use offers to whoever writes the input, such as an AI agent.

`lacre install-skill DIR` installs a skill that teaches an AI agent to use lacre:

```bash
lacre install-skill .claude/skills    # for Claude Code
lacre install-skill .agents/skills    # for Codex and other agents that read it
```

## Documentation

- [Usage](docs/usage.md): the page lacre builds, the options, the input, the style guide, the agent skill, exit codes
  and known limits
- [Themes](docs/themes.md): replacing the built-in page with your own layout, styles and fonts, and a style guide for
  whoever writes the input
- [Windows](docs/windows.md): `cmd.exe`, PowerShell and Git Bash
- [Development](docs/development.md): tasks, tests, CI and releases
