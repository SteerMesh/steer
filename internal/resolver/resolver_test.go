package resolver

import (
	"testing"
)

func TestResolveConstraint_exact(t *testing.T) {
	got, err := ResolveConstraint("1.0.0", []string{"1.0.0", "2.0.0"})
	if err != nil || got != "1.0.0" {
		t.Fatalf("got %q err %v", got, err)
	}
}

func TestResolveConstraint_caret(t *testing.T) {
	got, err := ResolveConstraint("^1.0.0", []string{"1.0.0", "1.1.0", "2.0.0"})
	if err != nil || got != "1.1.0" {
		t.Fatalf("got %q err %v", got, err)
	}
}

func TestResolveConstraint_tilde(t *testing.T) {
	got, err := ResolveConstraint("~1.0.0", []string{"1.0.0", "1.0.1", "1.1.0"})
	if err != nil || got != "1.0.1" {
		t.Fatalf("got %q err %v", got, err)
	}
}

func TestResolveConstraint_noMatch(t *testing.T) {
	_, err := ResolveConstraint("^2.0.0", []string{"1.0.0"})
	if err == nil {
		t.Fatal("expected error")
	}
}
