# The built-in theme

The built-in theme is plain HTML with a small stylesheet: no framework, no classes of its own and no JavaScript. Write
plain, semantic HTML; a class does nothing unless the input or a `--head` value brings the styles for it.

## What the page already has

The page is a complete document, so the input is only what goes inside `<body>`. Nothing visible goes around it, and
the title is only the `<title>` in the head, so a page that should show a heading starts the input with an `<h1>`.

## What it styles

The stylesheet fixes what browsers draw badly by default and adds little look of its own. Its rules are in a cascade
layer, so any style the input or a `--head` value brings wins over them, whatever its selectors and without
`!important`.

- **Text:** a centered column at most 48rem wide, in the system's sans-serif font at 112.5% of the browser's default
  size, usually 18 px. Long words and URLs break. Headings have the browser's sizes, so `<h5>` and `<h6>` are smaller
  than the text: use `<h1>` to `<h4>`.
- **Blocks:** lists and `<dd>` are indented by 2em, `<dt>` and `<summary>` are bold, and `<blockquote>`, `<hr>` and
  `<details>` get thin gray lines.
- **Code:** monospace at 0.875em on a light gray background, and `<kbd>` has a border. A nested one gets no box of its
  own. A long line in a `<pre>` scrolls, but wraps in print.
- **Tables:** as wide as the column, with a border and padding on every cell; one that should fit its content sets
  `width: auto`. `<code>` and links in a cell break anywhere, but other long text without spaces, such as a bare URL,
  makes the page scroll sideways.
- **Media and forms:** `<img>` and `<video>` shrink to fit the column and keep their proportions, and other media and
  form controls are never wider than it. Controls use the font of the text, and labels are bold, except those of
  checkboxes and radio buttons.

Links, `<mark>`, `<abbr>` and `<legend>` look as the browser draws them.

## Colors

The page declares no color scheme, so it is light, whatever the viewer's system setting: text `#1f1f1f` on `#f3f3f3`,
with translucent gray borders and backgrounds that show on any background, and the browser's colors for links and
`<mark>`. A page background goes on `body`: one on `html` leaves the theme's background on the column, as a stripe.

To follow a dark system setting, with text `#e5e5e5` on `#1f1f1f`, a `--head` value brings this. Browsers also apply
it at the start of the input, but there it is not valid HTML: a `<style>` in the body may hold only rules inside
`@scope`, which cannot reach `:root`.

```html
<style>:root { color-scheme: light dark }</style>
```

Do it only for input that sets no colors of its own, since they stay as they are on the dark page: the page's light
text vanishes on a white card, dark gray text hardly reads, and a diagram with black lines on a transparent background
disappears.

The colors are variables, which a `:root` rule can set: the four below, `--lacre-border` for every border and line, and
`--lacre-pre-bg` and `--lacre-code-bg` for the backgrounds of a `<pre>` and of code in the text. The two dark ones are
unused until the input or a `--head` value adds `color-scheme: light dark`.

```html
<style>
:root { --lacre-text-light: #000; --lacre-bg-light: #fff; --lacre-text-dark: #ddd; --lacre-bg-dark: #000 }
</style>
```
