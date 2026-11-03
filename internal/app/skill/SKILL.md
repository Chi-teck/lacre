---
name: lacre
description: >
  Use when a task ends in a standalone HTML page to open in a browser, share or print, such as a report or notes:
  write only the content, and the lacre command wraps it into a complete page in the user's theme. Not for HTML in a
  site's or app's source.
metadata:
  version: "{{version}}"
---

# Lacre

The lacre command wraps the content of a page, an HTML snippet, into a complete HTML page. Everything around the
content comes from a theme that the user picks: the document's structure, its head and its styles. So do not write
standalone HTML pages from scratch: write only the content, what the reader sees, and let lacre build the page.

HTML that is part of a website's or an app's source, such as its templates or components, is not for lacre: it follows
the project's own layout.

## Why not write pages from scratch

- **Cost:** a page's doctype, head and styles are often longer than its content, and writing is the slowest and
  costliest thing you do: output tokens cost more than input tokens and come one at a time. You read the theme's guide
  once, but would write a hand-made page's styles again for every page.
- **Context:** everything you write stays in your context for the rest of the session, so the boilerplate of each page
  crowds out the task. The page that lacre writes goes to a file and never enters your context.
- **Quality:** the head and styles come from a theme made once and improved over time, not made up for each page. Every
  page gets a complete head, with the charset, viewport and title, and your attention goes to the content, which is
  what the reader came for.
- **Consistency:** every page has the look of the user's theme, and the user restyles all their pages by changing the
  theme, without touching the content.

## Building a page

1. Run `lacre style-guide`. It prints the guide of the theme in use: what the content can use, such as a framework and
   its classes, and what the page already puts around the content. It is the only source on the theme, which the user
   can switch. If it fails with "no style guide", write plain HTML with no classes, and still build the page with lacre.
2. Write the content as an HTML snippet that uses what the guide offers, with no `<html>`, `<head>` or `<body>` of its
   own and nothing the page already has, such as a header. Pipe it in, or put it in a temporary file outside the
   user's project.
3. Build the page, with a `--title` and with `--lang` set to the language of the content:

   ```bash
   lacre build --title 'Quarterly report' --lang en > report.html <<'EOF'
   <p>Revenue grew by 12% over the quarter.</p>
   EOF
   ```

Pass `--theme` only when the user names a theme, and then to both commands, so that the guide describes the page.

The content goes into the page as it is, as HTML: the line breaks of plain text collapse, and its `<` and `&` are
read as markup. So escape plain text first, `&` to `&amp;` and then `<` to `&lt;`, and wrap it in paragraphs, or in a
`<pre>` when its line breaks and spacing matter, as in a log or code.

`lacre --help` describes every option.

## Adding styles and scripts

Add styles only for what the theme does not cover, such as a chart, or where the page needs something else, such as a
narrower table. The style guide may already offer what you need, such as a class or a variable, and tells how the
theme's styles give way to yours. Keep your styles few, and restyle the whole page only when the user asks: a look made
up for one page undoes the reasons above.

`--head` adds HTML at the end of the page's head and can be given more than once, as in
`--head '<style>.chart { height: 20rem }</style>'`. Put styles there, not in the content. Put a script at the end of
the content, or in `--head` as a `<script type="module">`: both run once the content is there. For more than a line,
write the HTML to a temporary file and pass `--head "$(cat head.html)"`.

Anything you add that loads by URL, such as a stylesheet, script, image or font, makes the page depend on it. For a
page that should be a single file that works offline, inline it instead: a library's code in `<style>` or `<script>`,
images and fonts as `data:` URLs. Put a large script, such as a library, at the end of the content rather than in
`--head`, since the system limits the size of a command-line argument.
