package app

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"
)

// wantSkillInstalled checks that lacre install-skill dir wrote the skill into
// file, with exit code 0, nothing on standard output and on standard error the
// line that names file.
func wantSkillInstalled(t *testing.T, dir, file string) {
	t.Helper()
	stdout, stderr, code := lacre("", "install-skill", dir)
	if code != 0 {
		t.Errorf("the exit code is %d, want 0", code)
	}
	if stdout != "" {
		t.Errorf("standard output is %q, want nothing", stdout)
	}
	if want := "lacre: wrote " + file + "\n"; stderr != want {
		t.Errorf("standard error is %q, want %q", stderr, want)
	}
	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != skillText() {
		t.Errorf("%s is not the skill:\n%q", file, b)
	}
}

// wantSkillError checks that lacre install-skill dir failed with exit code 1,
// nothing on standard output and on standard error "lacre: " followed by msg
// and sys, which are compared as wantThemeErrorArgs compares them.
func wantSkillError(t *testing.T, dir, msg, sys string) {
	t.Helper()
	stdout, stderr, code := lacre("", "install-skill", dir)
	if code != 1 {
		t.Errorf("the exit code is %d, want 1", code)
	}
	if stdout != "" {
		t.Errorf("standard output is %q, want nothing", stdout)
	}
	want := "lacre: " + msg
	if runtime.GOOS == "windows" {
		if !strings.HasPrefix(stderr, want) {
			t.Errorf("standard error is %q, want it to start with %q", stderr, want)
		}
		return
	}
	if want += sys + "\n"; stderr != want {
		t.Errorf("standard error is %q, want %q", stderr, want)
	}
}

// install-skill creates the directories it needs, replaces an existing skill
// and leaves the other files in the lacre directory alone. A relative
// directory is relative to the current directory, and the message names the
// file as filepath.Join gives it: cleaned, still relative and, on Windows,
// with \ between the names.
func TestInstallSkill(t *testing.T) {
	t.Run("new directories", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "project", "skills")
		wantSkillInstalled(t, dir, filepath.Join(dir, "lacre", "SKILL.md"))
	})
	t.Run("existing skill", func(t *testing.T) {
		dir := t.TempDir()
		skillDir := filepath.Join(dir, "lacre")
		if err := os.Mkdir(skillDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("An edited skill\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		const notes = "Notes\n"
		if err := os.WriteFile(filepath.Join(skillDir, "notes.md"), []byte(notes), 0o644); err != nil {
			t.Fatal(err)
		}
		wantSkillInstalled(t, dir, filepath.Join(skillDir, "SKILL.md"))
		b, err := os.ReadFile(filepath.Join(skillDir, "notes.md"))
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != notes {
			t.Errorf("notes.md is %q, want %q", b, notes)
		}
	})
	t.Run("relative directory", func(t *testing.T) {
		t.Chdir(t.TempDir())
		wantSkillInstalled(t, "./.agents/skills/", filepath.Join(".agents", "skills", "lacre", "SKILL.md"))
	})
	// Only a ~ that is a whole name is taken for one the shell did not expand.
	t.Run("name that starts with ~", func(t *testing.T) {
		t.Chdir(t.TempDir())
		wantSkillInstalled(t, "~skills", filepath.Join("~skills", "lacre", "SKILL.md"))
	})
}

// A directory that cannot be created or a skill that cannot be written is one
// line that names the path, with no usage after it. The directory named is the
// one that failed, which need not be lacre.
func TestInstallSkillErrors(t *testing.T) {
	t.Run("file in the way", func(t *testing.T) {
		dir := t.TempDir()
		file := filepath.Join(dir, "file")
		if err := os.WriteFile(file, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		wantSkillError(t, filepath.Join(file, "skills"), "create directory "+file+": ", "not a directory")
	})
	t.Run("lacre is a file", func(t *testing.T) {
		dir := t.TempDir()
		skillDir := filepath.Join(dir, "lacre")
		if err := os.WriteFile(skillDir, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		wantSkillError(t, dir, "create directory "+skillDir+": ", "not a directory")
	})
	t.Run("SKILL.md is a directory", func(t *testing.T) {
		dir := t.TempDir()
		file := filepath.Join(dir, "lacre", "SKILL.md")
		if err := os.MkdirAll(file, 0o755); err != nil {
			t.Fatal(err)
		}
		wantSkillError(t, dir, "write "+file+": ", "is a directory")
	})
	t.Run("read-only directory", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Chmod(dir, 0o555); err != nil {
			t.Fatal(err)
		}
		// Windows cannot remove a read-only directory, so it is made
		// writable again before the temporary directory goes.
		t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
		skillDir := filepath.Join(dir, "lacre")
		if err := os.Mkdir(skillDir, 0o755); err == nil {
			t.Skip("the directory can still be written, as on Windows or as root")
		}
		wantSkillError(t, dir, "create directory "+skillDir+": ", "permission denied")
	})
}

// The skill's front matter, read by hand, since the tests have no YAML parser:
// it is between two --- lines at the start, names the skill lacre, as the
// directory install-skill writes it to, and has a description of 1 to 1024
// characters, the limit of the Agent Skills specification. The description is
// the indented lines after "description: >", so a line that is not indented
// ends it early, where Claude Code would drop it and take the first heading
// instead. skillText puts the version in place of the placeholder.
func TestSkill(t *testing.T) {
	text := skillText()
	lines := strings.Split(text, "\n")
	end := slices.Index(lines[1:], "---") + 1
	if lines[0] != "---" || end == 0 {
		t.Fatalf("the skill does not start with front matter between --- lines:\n%s", text)
	}
	front := lines[1:end]
	if !slices.Contains(front, "name: lacre") {
		t.Error("the front matter has no line name: lacre")
	}
	i := slices.Index(front, "description: >")
	if i < 0 {
		t.Fatal("the front matter has no line description: >")
	}
	var words []string
	for _, line := range front[i+1:] {
		after, ok := strings.CutPrefix(line, "  ")
		if !ok {
			break
		}
		words = append(words, after)
	}
	description := strings.Join(words, " ")
	if n := utf8.RuneCountInString(description); n == 0 || n > 1024 {
		t.Errorf("the description has %d characters, want 1 to 1024: %q", n, description)
	}
	if want := `  version: "` + buildVersion() + `"`; !slices.Contains(front, want) {
		t.Errorf("the front matter has no line %q", want)
	}
	if strings.Contains(text, "{{version}}") {
		t.Error("the skill still holds {{version}}")
	}
}

// The skill holds no carriage return, so it has LF line endings even in a
// program built from a Windows checkout. .gitattributes keeps them out; this
// catches a file that got them anyway.
func TestSkillNoCR(t *testing.T) {
	if strings.ContainsRune(skillFile, '\r') {
		t.Error("skill/SKILL.md holds a carriage return")
	}
}
