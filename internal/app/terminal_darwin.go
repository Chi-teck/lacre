package app

import (
	"syscall"
	"unsafe"
)

// stdinIsTerminal says whether standard input is a terminal. The
// standard library has no such check, and "is it a character device?" would
// refuse /dev/null too. So this asks for the terminal settings with the
// TIOCGETA ioctl, as golang.org/x/term does: only a terminal has them.
// golang.org/x/term reaches ioctl through libc, while syscall.Syscall on macOS
// traps into the kernel directly, which is the only way the standard library
// offers.
func stdinIsTerminal() bool {
	var t syscall.Termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(syscall.Stdin), syscall.TIOCGETA,
		uintptr(unsafe.Pointer(&t)))
	return errno == 0
}
