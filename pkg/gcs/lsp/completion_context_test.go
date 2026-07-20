package lsp

import (
	"strings"
	"testing"

	"go.lsp.dev/protocol"
)

func TestCompleteWeaponStringContext(t *testing.T) {
	text := `mualani add weapon="wid`
	// cursor at end
	line, char := endPos(text)
	items := mustCompletionItems(t, Complete(text, line, char))
	if len(items) == 0 {
		t.Fatal("expected weapon completions")
	}
	for _, it := range items {
		d, _ := it.Detail.Get()
		if d != "weapon" {
			t.Fatalf("expected only weapons, got %q (%s)", it.Label, d)
		}
	}
	// widsith should be present for prefix wid
	if !hasLabel(items, "widsith") && !hasLabelPrefix(items, "wid") {
		t.Fatalf("expected weapon matching wid*, got %v", labels(items))
	}
}

func TestCompleteSetStringContext(t *testing.T) {
	text := `mualani add set="obs`
	line, char := endPos(text)
	items := mustCompletionItems(t, Complete(text, line, char))
	if len(items) == 0 {
		t.Fatal("expected set completions")
	}
	for _, it := range items {
		d, _ := it.Detail.Get()
		if d != "artifact set" {
			t.Fatalf("expected only sets, got %q (%s)", it.Label, d)
		}
	}
}

func TestCompleteActiveCharacterContext(t *testing.T) {
	text := "active mav"
	line, char := endPos(text)
	items := mustCompletionItems(t, Complete(text, line, char))
	if len(items) == 0 {
		t.Fatal("expected character completions")
	}
	for _, it := range items {
		d, _ := it.Detail.Get()
		if d != "character" {
			t.Fatalf("expected only characters, got %q (%s)", it.Label, d)
		}
	}
}

func TestCompleteOptionsContext(t *testing.T) {
	text := "options swap"
	line, char := endPos(text)
	items := mustCompletionItems(t, Complete(text, line, char))
	if len(items) == 0 {
		t.Fatal("expected option completions")
	}
	for _, it := range items {
		d, _ := it.Detail.Get()
		if d != "option" {
			t.Fatalf("expected only options, got %q (%s)", it.Label, d)
		}
	}
}

func TestCompleteFieldContext(t *testing.T) {
	text := "while .mualani.night"
	line, char := endPos(text)
	items := mustCompletionItems(t, Complete(text, line, char))
	if len(items) == 0 {
		t.Fatal("expected field completions")
	}
	for _, it := range items {
		d, _ := it.Detail.Get()
		if d != "field" {
			t.Fatalf("expected only fields, got %q (%s)", it.Label, d)
		}
	}
}

func TestCompleteActionLineContext(t *testing.T) {
	text := "mualani ski"
	line, char := endPos(text)
	items := mustCompletionItems(t, Complete(text, line, char))
	if len(items) == 0 {
		t.Fatal("expected action completions")
	}
	for _, it := range items {
		d, _ := it.Detail.Get()
		if d != "action" {
			t.Fatalf("expected only actions, got %q (%s)", it.Label, d)
		}
	}
}

func TestCompleteStatsContext(t *testing.T) {
	text := "mualani add stats hp"
	line, char := endPos(text)
	items := mustCompletionItems(t, Complete(text, line, char))
	if len(items) == 0 {
		t.Fatal("expected stat completions")
	}
	for _, it := range items {
		d, _ := it.Detail.Get()
		if d != "stat" {
			t.Fatalf("expected only stats, got %q (%s)", it.Label, d)
		}
	}
}

func TestCompleteUnscopedStillBroad(t *testing.T) {
	text := "xia"
	line, char := endPos(text)
	items := mustCompletionItems(t, Complete(text, line, char))
	if len(items) == 0 {
		t.Fatal("expected completions")
	}
	// should include characters (xiangling etc.), not only one category
	var kinds []string
	seen := map[string]bool{}
	for _, it := range items {
		d, _ := it.Detail.Get()
		if !seen[d] {
			seen[d] = true
			kinds = append(kinds, d)
		}
	}
	if !seen["character"] {
		t.Fatalf("expected characters in unscoped complete, kinds=%v", kinds)
	}
}

func TestContextWeaponDetect(t *testing.T) {
	text := `x add weapon="`
	scope, filter := completionContext(text, 0, uint32(len(text)))
	if scope != scopeWeapon {
		t.Fatalf("scope=%v want weapon, filter=%q", scope, filter)
	}
}

func endPos(text string) (line, char uint32) {
	lines := strings.Split(text, "\n")
	line = uint32(len(lines) - 1)
	char = uint32(len(lines[len(lines)-1]))
	return line, char
}

func hasLabel(items []protocol.CompletionItem, label string) bool {
	for _, it := range items {
		if it.Label == label {
			return true
		}
	}
	return false
}

func hasLabelPrefix(items []protocol.CompletionItem, prefix string) bool {
	for _, it := range items {
		if strings.HasPrefix(strings.ToLower(it.Label), prefix) {
			return true
		}
	}
	return false
}

func labels(items []protocol.CompletionItem) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = it.Label
	}
	return out
}
