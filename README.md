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

## Documentation

- [Usage](docs/usage.md): the page lacre builds, the options, the input, exit codes and known limits
- [Themes](docs/themes.md): replacing the built-in page with your own layout, styles and fonts
- [Windows](docs/windows.md): `cmd.exe`, PowerShell and Git Bash
- [Development](docs/development.md): tasks, tests, CI and releases
