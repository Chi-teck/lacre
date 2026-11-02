//go:build unix

package app

import (
	"os"
	"syscall"
)

// stdinIsTerminal says whether standard input is a terminal. The standard
// library has no such check. golang.org/x/term asks for the terminal settings
// with an ioctl, but on macOS the standard library can make that call only by
// trapping into the kernel directly, which Apple does not support. So this
// uses only Stat, which reaches the kernel the supported way everywhere. A
// terminal is a character device, and so is /dev/null, which the device
// number tells apart: the number rather than the file, since a container or a
// chroot can have a /dev/null node of its own. Other character devices, such
// as /dev/zero, pass for terminals, so lacre refuses them with the hint
// instead of reading them without end. A failed Stat says "not a terminal",
// and lacre reads the way cat does.
func stdinIsTerminal() bool {
	in, err := os.Stdin.Stat()
	if err != nil || in.Mode()&os.ModeCharDevice == 0 {
		return false
	}
	null, err := os.Stat(os.DevNull)
	if err != nil {
		return false
	}
	return in.Sys().(*syscall.Stat_t).Rdev != null.Sys().(*syscall.Stat_t).Rdev
}
