package app

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// T-15. The binary uses only the standard library (principle 4), and the
// test-only golang.org/x/net never reaches it (D1). go test runs this in
// internal/app, so the command names the main package by its path.
func TestStandardLibraryOnly(t *testing.T) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go is not on PATH")
	}
	// go test reuses a cached result unless a file this process read has
	// changed; what go list reads does not count. main.go is the binary's
	// only file outside internal/app, so reading it makes a change there run
	// the test again.
	if _, err := os.ReadFile(filepath.Join("..", "..", "main.go")); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(goBin, "list", "-deps", "-f", "{{if not .Standard}}{{.ImportPath}}{{end}}",
		"github.com/Chi-teck/lacre")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list: %v\n%s", err, stderr.Bytes())
	}
	got := strings.Fields(string(out))
	slices.Sort(got)
	want := []string{"github.com/Chi-teck/lacre", "github.com/Chi-teck/lacre/internal/app"}
	if !slices.Equal(got, want) {
		t.Errorf("the binary's packages outside the standard library are %q, want only %q", got, want)
	}
}
