# The built-in theme

The built-in page is plain HTML with a small stylesheet: no framework, no classes of its own and no JavaScript. Write
plain, semantic HTML; a class does nothing unless the input or a `--head` value brings the styles for it.

## What the page already has

The page is a complete document. Its head has the charset, the viewport tag, the title and the stylesheet, so the input
is only what goes inside `<body>`, not a document with an `<html>`, `<head>` or `<body>` of its own.

Nothing visible goes around the input: no header, navigation, heading or footer. The title is only the `<title>` in the
head, so a page that should show a heading starts the input with an `<h1>`.

## What it styles

The stylesheet uses element selectors only, so any style the input or a `--head` value brings wins over it.

- `body`: a centred column at most 48rem wide, with 1rem of padding, the system's sans-serif font and a line height of
  1.5. A long word or URL breaks instead of running past the edge.
- `pre`: the system's monospace font at 0.875em; long lines wrap instead of being cut off.
- `img` and `video`: they shrink to fit the column and keep their proportions.

Everything else, such as headings, lists, tables, links, forms and `<code>` outside a `<pre>`, looks as the browser
draws it. A table has no borders and almost no space between cells.

## Colours

The page sets no colours: it has the browser's own, dark text on a light background, whatever the viewer's system
setting. An input that needs colours, or a dark page, brings them in a `<style>` element given with `--head`, or in
`style` attributes. A `<style>` at the start of the input works too: browsers apply it, although HTML allows it only in
the head.

## What to avoid

- Plain text as it is: its line breaks collapse into one paragraph, and `<` and `&` are read as markup. Escape it and
  put it in a `<pre>`, or mark it up as paragraphs.
- Stylesheets, scripts, images and fonts from elsewhere, when the page should be a single file that works offline:
  nothing in the page refers outside it unless the input does.
