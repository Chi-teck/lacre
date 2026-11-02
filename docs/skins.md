# Skins

A skin replaces the built-in page with your own: your layout, a framework such as Bootstrap, your own fonts or a dark
theme. It is a directory that holds a `page.gohtml` and the files it uses. Set `LACRE_SKIN` to the directory, and
`lacre build` uses it:

```bash
export LACRE_SKIN="$HOME/skins/my-skin"
echo 'Hello world!' | lacre build > hello-world.html
```

A relative path is relative to the current directory, so a shell profile should set an absolute one. When `LACRE_SKIN`
is unset or empty, lacre uses its built-in skin. Nothing on the command line shows the skin, so the same command can
give a different page where the variable is not set, such as in a cron job or in CI.

```
my-skin/
├── page.gohtml        the template; the only file a skin must have
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
| `.Language` | The `--lang` value, escaped for HTML, or nothing; the skin decides what to write without one                    |
| `.Head`     | Each non-empty `--head` value, as it is, followed by a line break; nothing without them                         |
| `.Body`     | `<!-- lacre:begin -->`, a line break, the input as it is, and `<!-- lacre:end -->`, with no line break after it |

Two functions bring the skin's other files into the page, so that a skinned page can still be a single file that works
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

The built-in skin, in [`internal/app/skin/`](../internal/app/skin), is a working example: its `page.gohtml` and
`style.css` give the page lacre builds without a skin.

- **Paths** in `include` and `dataURI` are names joined by `/` on every system, Windows included: `fonts/inter.woff2`,
  never `fonts\inter.woff2`. No name may be empty, `.` or `..`, and a path may not hold `\` or `:`.
- **A skin cannot read files outside its directory,** not even through a symbolic link. Skins get shared, and a
  downloaded skin must not be able to copy a private file, such as an SSH key, into a page that you then publish.
- **Included files are not templates:** a `{{` in `style.css` is plain text, so a font's `url(…)` goes in `page.gohtml`,
  as in the example. The other way round, a `{{` that `page.gohtml` needs as text, as in some inline JavaScript, goes in
  an included file.
- **`.Title` and `.Language` are escaped for HTML,** in text and in attributes, not for JavaScript. A skin that puts the
  title inside a `<script>` has to deal with that itself.
- **The line break after `<!-- lacre:begin -->` is part of `.Body`,** so a skin that puts `.Body` straight inside a
  `<pre>` shows an empty first line.
- **lacre does not check that a skin uses the values.** A skin without `.Head` drops the `--head` values, and one
  without `.Body` drops the input.
- **Only the built-in skin promises a self-contained page.** A skin that links to a stylesheet on a CDN makes its pages
  depend on it.
- **A broken skin stops lacre before it writes anything,** with exit code 2 and a message that names the skin and the
  file, and the line for a template error: a missing directory or `page.gohtml`, a template error, or a file that
  `include` or `dataURI` cannot read.

The type in a `data:` URI comes from the file's extension, in upper or lower case, through this fixed list, never from
the system's settings, so that the same skin gives the same bytes on every machine. Any other extension is an error.

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

To build one page with the built-in skin while `LACRE_SKIN` is set, empty the variable for that command. In Unix
shells:

```bash
LACRE_SKIN= lacre build < snippet.html > page.html
```

`cmd.exe` and PowerShell have no form for one command. In `cmd.exe`, `set LACRE_SKIN=` lasts for the rest of the
session, and this covers one command:

```bat
cmd /c "set LACRE_SKIN=& lacre build < snippet.html > page.html"
```

The `&` comes right after the `=`: a space before it sets the variable to a space, which lacre takes for a skin path. In
PowerShell, `Remove-Item Env:LACRE_SKIN` lasts for the rest of the session. These Windows forms have not been checked on
a Windows machine yet.
