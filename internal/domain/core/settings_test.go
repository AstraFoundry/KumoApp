package core

import (
	"encoding/json"
	"testing"
)

func TestTUNStackDefaultsToMips(t *testing.T) {
	if got := DefaultTUNSettings().Stack; got != "mips" {
		t.Fatalf("default stack %q", got)
	}
	var s TUNSettings
	if err := json.Unmarshal([]byte(`{"isEnabled":true}`), &s); err != nil {
		t.Fatal(err)
	}
	if s.Stack != "mips" {
		t.Fatalf("a missing stack became %q", s.Stack)
	}
}

func TestStoredTUNStackSurvives(t *testing.T) {
	for _, stack := range TUNStacks {
		var s TUNSettings
		if err := json.Unmarshal([]byte(`{"stack":"`+stack+`"}`), &s); err != nil {
			t.Fatal(err)
		}
		if got := s.Normalized().Stack; got != stack {
			t.Errorf("stored %q became %q", stack, got)
		}
	}
	if got := (TUNSettings{Stack: "lwip"}).Normalized().Stack; got != "mips" {
		t.Errorf("an unknown stack became %q", got)
	}
}
