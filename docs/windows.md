# Windows

lacre works in `cmd.exe`, in PowerShell and in Unix shells such as Git Bash. The examples in the README and the
other documents are for Unix shells.

In `cmd.exe`, `<` and `>` work as in Unix shells. Quote values with double quotes, and leave the quotes out of `echo`,
or they end up in the page:

```bat
echo Hello world! | lacre build > hello-world.html
lacre build --title "My page" --lang en < snippet.html > page.html
```

The file form of `--head`, `--head "$(cat head.html)"`, is for Unix shells: `cmd.exe` has no command substitution and
would pass `$(cat head.html)` to lacre as text, which then ends up in the page. In `cmd.exe`, give `--head` the HTML
itself, or run lacre in Git Bash.

In PowerShell, `echo 'Hello world!' | lacre build > hello-world.html` works too, but PowerShell has no `<`.

Instead, pipe the file in. This works, but PowerShell may change the line endings and add a line break at the end:

```powershell
Get-Content snippet.html -Raw | lacre build > page.html
```

To pass the bytes exactly, let `cmd.exe` do the redirection:

```powershell
cmd /c "lacre build < snippet.html > page.html"
```

Known limit: Windows PowerShell 5.1 can mangle non-ASCII text in pipes and in `>`. PowerShell 7.4 and later and
`cmd /c` do not. PowerShell before 7.3 also drops double quotes inside arguments passed to programs, which breaks
`--head` values with attributes, such as `--head '<meta name="description" content="A test page" />'`.

The PowerShell details on this page have not been checked on a Windows machine yet.
