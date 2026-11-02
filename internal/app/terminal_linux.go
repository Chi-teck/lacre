package app

import (
	"syscall"
	"unsafe"
)

// stdinIsTerminal says whether standard input is a terminal (CLI-11). The
// standard library has no such check, and "is it a character device?" would
// refuse /dev/null too. So this asks for the terminal settings with the
// TCGETS ioctl, as golang.org/x/term does: only a terminal has them.
func stdinIsTerminal() bool {
	var t syscall.Termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(syscall.Stdin), syscall.TCGETS,
		uintptr(unsafe.Pointer(&t)))
	return errno == 0
}
