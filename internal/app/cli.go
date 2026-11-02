// Package app is lacre itself. The root main.go is a shim over Main below, and
// carries the command's own documentation, where `go doc` looks for it.
package app

import "fmt"

// Main is the command. For now it only prints the version line,
// "lacre <version>".
func Main() {
	fmt.Printf("lacre %s\n", buildVersion())
}
