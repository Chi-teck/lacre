# Windows

lacre works in `cmd.exe`, in PowerShell and in Unix shells such as Git Bash. The examples in the README and the other
documents are for Unix shells; this page shows what differs in `cmd.exe` and PowerShell. The PowerShell details have
not been checked on a Windows machine yet.

## cmd.exe

`<` and `>` work as in Unix shells. Quote values with double quotes, and leave the quotes out of `echo`, or they end up
in the page:

```bat
echo Hello world! | lacre build > hello-world.html
lacre build --title "My page" --lang en < snippet.html > page.html
```

The file form of `--head`, `--head "$(cat head.html)"`, is for Unix shells: `cmd.exe` has no command substitution and
would pass `$(cat head.html)` to lacre as text, which then ends up in the page. In `cmd.exe`, give `--head` the HTML
itself, or run lacre in Git Bash.

## PowerShell

`echo 'Hello world!' | lacre build > hello-world.html` works as in Unix shells, but PowerShell has no `<`. Pipe the file
in instead. This works, but PowerShell may change the line endings and add a line break at the end:

```powershell
Get-Content snippet.html -Raw | lacre build > page.html
```

To pass the bytes exactly, let `cmd.exe` do the redirection:

```powershell
cmd /c "lacre build < snippet.html > page.html"
```

Older versions have known limits:

- Windows PowerShell 5.1 can mangle non-ASCII text in pipes and in `>`. PowerShell 7.4 and later and `cmd /c` do not.
- PowerShell before 7.3 drops double quotes inside arguments passed to programs, which breaks `--head` values with
  attributes, such as `--head '<meta name="description" content="A test page" />'`. It drops empty arguments, such as
  `''`, too, so an empty `--theme` is written `--theme=`.

## Installing the agent skill

`lacre install-skill` takes Windows paths, and its success line names the file with `\`, as in
`lacre: wrote .agents\skills\lacre\SKILL.md`. `cmd.exe` and PowerShell before 7.6 do not expand `~`, so lacre rejects
`lacre install-skill ~\.claude\skills` there with exit code 2. Name the home directory instead; this works in every
version. In `cmd.exe`:

```bat
lacre install-skill "%USERPROFILE%\.claude\skills"
```

In PowerShell:

```powershell
lacre install-skill "$HOME\.claude\skills"
```

lacre writes the skill itself, so the limit of Windows PowerShell 5.1 on `>` does not affect it.
