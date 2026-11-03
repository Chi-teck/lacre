# Themes

A theme replaces the built-in page with your own: your layout, a framework such as Bootstrap, your own fonts or a dark
color scheme. It is a directory that holds a `page.gohtml` and the files it uses. Name the directory with `--theme`:

```bash
lacre build --theme "$HOME/themes/my-theme" < snippet.html > page.html
```

Or set `LACRE_THEME` to it, and `lacre build` uses it without `--theme`:

```bash
export LACRE_THEME="$HOME/themes/my-theme"
echo 'Hello world!' | lacre build > hello-world.html
```

A relative path is relative to the current directory, so a shell profile should set an absolute one. `--theme` wins
over `LACRE_THEME`, and given more than once, the last one wins. An empty `--theme`, or no `--theme` and an unset or
empty `LACRE_THEME`, gives the built-in theme. With only the variable, nothing on the command line shows the theme, so
the same command can give a different page where the variable is not set, such as in a cron job or in CI: a script
that needs a given theme should name it with `--theme`.

```
my-theme/
├── page.gohtml        the template; the only file a theme must have
├── style-guide.md     what to write for the theme; optional
├── style.css
├── logo.svg
└── fonts/
    └── inter.woff2
```

`page.gohtml` is the whole page, written in Go's [template syntax](https://pkg.go.dev/text/template). It gets four
values:

| Value       | What it holds                                                                                                   |
|-------------|-----------------------------------------------------------------------------------------------------------------|
| `.Title`    | The `--title` value, or `Untitled`, escaped for HTML                                                            |
| `.Language` | The `--lang` value, escaped for HTML, or nothing; the theme decides what to write without one                   |
| `.Head`     | Each non-empty `--head` value, as it is, followed by a line break; nothing without them                         |
| `.Body`     | `<!-- lacre:begin -->`, a line break, the input as it is, and `<!-- lacre:end -->`, with no line break after it |

Two functions bring the theme's other files into the page, so that a themed page can still be a single file that works
offline:

- `include "path"` gives the file's content, as it is: CSS for a `<style>`, JavaScript for a `<script>`, an SVG or an
  HTML fragment.
- `dataURI "path"` gives the file as a `data:` URI, `data:<type>;base64,<data>`, for images and fonts: `<img src="…">`,
  or `url(…)` in a `@font-face`. The type comes from the file's extension, through the list below.

Everything else in `page.gohtml` is copied as written, HTML comments included; a template comment, `{{/* … */}}`, leaves
nothing. The names of the values are case-sensitive, and any other name, such as `.title`, is an error. For example,
with the files above:

```html
<!DOCTYPE html>
<html{{if .Language}} lang="{{.Language}}"{{end}}>
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<title>{{.Title}}</title>
<style>
@font-face { font-family: Inter; src: url({{dataURI "fonts/inter.woff2"}}); }
{{include "style.css"}}</style>
{{.Head}}</head>
<body>
<header><img src="{{dataURI "logo.svg"}}" alt="" /></header>
<main>
{{.Body}}
</main>
</body>
</html>
```

The built-in theme, in [`internal/app/theme/`](../internal/app/theme), is a working example: its `page.gohtml` and
`style.css` give the page lacre builds without a theme, and its `style-guide.md` describes that page.

- **Paths** in `include` and `dataURI` are names joined by `/` on every system, Windows included: `fonts/inter.woff2`,
  never `fonts\inter.woff2`. No name may be empty, `.` or `..`, and a path may not hold `\` or `:`.
- **A theme cannot read files outside its directory,** not even through a symbolic link. Themes get shared, and a
  downloaded theme must not be able to copy a private file, such as an SSH key, into a page that you then publish.
- **Included files are not templates:** a `{{` in `style.css` is plain text, so a font's `url(…)` goes in `page.gohtml`,
  as in the example. The other way round, a `{{` that `page.gohtml` needs as text, as in some inline JavaScript, goes in
  an included file.
- **`.Title` and `.Language` are escaped for HTML,** in text and in attributes, not for JavaScript. A theme that puts
  the title inside a `<script>` has to deal with that itself.
- **The line break after `<!-- lacre:begin -->` is part of `.Body`,** so a theme that puts `.Body` straight inside a
  `<pre>` shows an empty first line.
- **lacre does not check that a theme uses the values.** A theme without `.Head` drops the `--head` values, and one
  without `.Body` drops the input.
- **Only the built-in theme promises a self-contained page.** A theme that links to a stylesheet on a CDN makes its
  pages depend on it.
- **A broken theme stops lacre before it writes anything,** with exit code 2 and a message that names the theme and the
  file, and the line for a template error: a missing directory or `page.gohtml`, a template error, or a file that
  `include` or `dataURI` cannot read.

The type in a `data:` URI comes from the file's extension, in upper or lower case, through this fixed list, never from
the system's settings, so that the same theme gives the same bytes on every machine. Any other extension is an error.

| Extension       | Type                       |
|-----------------|----------------------------|
| `.png`          | `image/png`                |
| `.jpg`, `.jpeg` | `image/jpeg`               |
| `.gif`          | `image/gif`                |
| `.webp`         | `image/webp`               |
| `.avif`         | `image/avif`               |
| `.svg`          | `image/svg+xml`            |
| `.ico`          | `image/vnd.microsoft.icon` |
| `.bmp`          | `image/bmp`                |
| `.woff`         | `font/woff`                |
| `.woff2`        | `font/woff2`               |
| `.ttf`          | `font/ttf`                 |
| `.otf`          | `font/otf`                 |

`.ico` has two types in common use: `image/vnd.microsoft.icon` is the registered one, and browsers accept both.

To build one page with the built-in theme while `LACRE_THEME` is set, give an empty `--theme`:

```bash
lacre build --theme= < snippet.html > page.html
```

`--theme=` is the same in every shell, `cmd.exe` and PowerShell included; write it rather than `--theme ''`, which
older PowerShell drops (see [Windows](windows.md)). This has not been checked on a Windows machine yet.

## Style guide

A theme can hold a `style-guide.md` next to its `page.gohtml`: a Markdown file in which you tell whoever writes the
input, often an AI agent, how to write for your theme. `lacre style-guide` prints it as it is, for the theme that
`--theme` or `LACRE_THEME` picks, as `lacre build` does:

```bash
lacre style-guide --theme "$HOME/themes/my-theme"
```

The file is not a template: a `{{` in it is printed as written, so a guide can quote template syntax. It is optional,
and a theme without one still builds pages. Only `lacre style-guide` fails then, with exit code 2 and this line on
standard error, rather than with stand-in text that would pass a guess off as your word:

```
lacre: theme /home/me/themes/my-theme: no style guide: style-guide.md does not exist
```

`style-guide.md` is read like the theme's other files, from its directory only. A symbolic link that leads out of the
theme, a `style-guide.md` that is a directory or one that cannot be read fails as a broken theme does. The command also
reads and parses `page.gohtml`, and fails if the template does not parse, so that nobody writes input for a page that
cannot be built. It does not run the template, so it does not check the files that `include` and `dataURI` read.

There is no required format. A guide could cover:

- **The framework and its exact version,** first, if the theme brings one: Bootstrap 5.3.3, not just Bootstrap. An
  agent knows the common frameworks already and often cannot follow a link, being offline or without a web tool, so the
  name and version tell it more than a link would. A link to the framework's documentation is for people, and serves
  them better than a list of every class.
- **The theme's own classes,** and what each one is for.
- **What the page already puts around the input,** such as a header, navigation or a heading, so that the input does
  not repeat it.
- **The color scheme,** such as a dark background, so that the input does not bring colors that clash with it.
- **What the input should avoid,** such as markup that the theme's styles do not handle.

The built-in theme's [`style-guide.md`](../internal/app/theme/style-guide.md) is an example.

**A style guide is text an agent acts on,** written by whoever wrote the theme. A downloaded theme can already put any
script into its pages, but its style guide can also give instructions to an agent that has a shell. Read the style
guide of a downloaded theme before an agent does.
