# Usage

`lacre build` reads all of standard input and writes the page to standard output. Messages go to standard error.

```bash
lacre build --title 'My page' --lang en < snippet.html > page.html
```

The page is a complete HTML document. Its head has the UTF-8 charset, a viewport tag so that the page fits a phone's
screen, a generator tag that names lacre, the title (`Untitled` without `--title`) and a small built-in stylesheet,
which gives a centred column that is easy to read, the system fonts, long lines in `<pre>` that wrap, and images and
videos that shrink to fit the column. Its body is your input, as it is, between two markers. That is the page of the
built-in theme, and a [theme](themes.md) of your own can change all of it. With the built-in theme, nothing lacre adds
refers to anything outside the page, so the page is a single file that works offline, unless what you put in it refers
to outside files. The same input, options and theme always give the same bytes.

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

`--head` adds HTML at the end of `<head>`, as it is, and can be given more than once. It comes after the built-in
stylesheet, so a `<style>` added this way overrides its rules:

```bash
lacre build --head '<style>body { max-width: 60rem }</style>' < snippet.html > page.html
```

To add a file, let the shell read it:

```bash
lacre build --head "$(cat head.html)" < snippet.html > page.html
```

A `--head` value that refers to outside files, such as a `<link>` to a stylesheet or a `<script src>`, makes the page
depend on them.

## Options

The options of `build`:

| Option         | Value      | Default    | Effect                                                                                       |
|----------------|------------|------------|----------------------------------------------------------------------------------------------|
| `--title TEXT` | any string | `Untitled` | The page title                                                                               |
| `--lang CODE`  | any string | none       | The page language, the `lang` attribute of `<html>`; without it there is no `lang` attribute |
| `--head HTML`  | any string | none       | HTML added at the end of `<head>`, as it is; can be repeated                                 |
| `--help`, `-h` | none       |            | Print the usage to standard output and exit with 0                                           |

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

## Exit codes

| Code | When                                                                                                                                                       |
|------|------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0    | The page was written, or the usage or version was printed                                                                                                  |
| 1    | Reading standard input or writing standard output failed                                                                                                   |
| 2    | The tool was called wrongly: no command, an unknown command, a bad option, an argument, a terminal or another character device as input, or a broken theme |

## Known limits

When standard input is a terminal, `lacre build` does not wait for typed input: it prints a hint and exits with 2. On
Linux, macOS and other Unix systems, lacre takes any character device other than `/dev/null` for a terminal, so it
refuses `/dev/zero` or `/dev/urandom` the same way. In Git Bash (mintty), lacre cannot tell that standard input is a
terminal, so it waits for input, as `cat` does. Ctrl+D ends the input. On systems that are neither Unix nor Windows,
such as Plan 9, lacre does not check at all and waits the same way.
