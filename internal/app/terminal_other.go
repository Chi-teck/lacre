//go:build !linux && !darwin && !windows

package app

// stdinIsTerminal says "not a terminal" on every other system. There is no
// check here, so lacre reads a terminal the way cat does, and go install
// still builds.
func stdinIsTerminal() bool { return false }
