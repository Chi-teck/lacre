//go:build unix

package app

import (
	"os"
	"path/filepath"
	"testing"
)

// The real check, with os.Stdin swapped for each input. /dev/zero is the case
// where it says "terminal": a character device, as a terminal is, but not
// /dev/null, so only the device number tells the two apart. CI has no
// terminal to give the check, so this is what shows it can say yes at all.
func TestStdinIsTerminal(t *testing.T) {
	file := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(file, []byte("Hello world!\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	open := func(name string) *os.File {
		f, err := os.Open(name)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = f.Close() })
		return f
	}
	pipe, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pipe.Close(); _ = w.Close() })

	cases := []struct {
		name  string
		stdin *os.File
		want  bool
	}{
		{"os.DevNull", open(os.DevNull), false},
		{"/dev/zero", open("/dev/zero"), true},
		{"file", open(file), false},
		{"pipe", pipe, false},
	}
	saved := os.Stdin
	t.Cleanup(func() { os.Stdin = saved })
	for _, c := range cases {
		os.Stdin = c.stdin
		if got := stdinIsTerminal(); got != c.want {
			t.Errorf("%s: stdinIsTerminal() is %v, want %v", c.name, got, c.want)
		}
	}
}
