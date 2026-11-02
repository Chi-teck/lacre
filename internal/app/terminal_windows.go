package app

import "syscall"

// stdinIsTerminal says whether standard input is a console, with
// GetConsoleMode, as golang.org/x/term does. It fails for anything else: a
// pipe, a file, NUL. In Git Bash, mintty gives programs a pipe, so a terminal
// there is not seen, a limit the README names.
func stdinIsTerminal() bool {
	var mode uint32
	return syscall.GetConsoleMode(syscall.Stdin, &mode) == nil
}
