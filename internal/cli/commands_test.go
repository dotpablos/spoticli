package cli

import "testing"

func TestFindCommandMatchesAlias(t *testing.T) {
	cmd, ok := findCommand("--help")
	if !ok {
		t.Fatal("expected help alias to be found")
	}
	if cmd.name != "help" {
		t.Fatalf("got %q, want help", cmd.name)
	}
}

func TestRequirePositiveCountArg(t *testing.T) {
	got, err := requirePositiveCountArg([]string{"3"})
	if err != nil {
		t.Fatalf("requirePositiveCountArg() error = %v", err)
	}
	if got != 3 {
		t.Fatalf("requirePositiveCountArg() = %d, want 3", got)
	}
}
