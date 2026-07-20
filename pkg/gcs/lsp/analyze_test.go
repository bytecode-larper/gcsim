package lsp

import (
	"strings"
	"testing"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestAnalyzeValidConfig(t *testing.T) {
	src := `
options iteration=1 duration=10 swap_delay=12;
xiangling char lvl=90/90 cons=0 talent=9,9,9;
xiangling add weapon="thecatch" refine=5 lvl=90/90;
xiangling add set="emblemofseveredfate" count=4;
xiangling add stats hp=4780 atk=311 er=0.518 pyro%=0.466 cr=0.311;
active xiangling;
target lvl=100 resist=0.1;
xiangling skill;
`
	diags := Analyze(src)
	for _, d := range diags {
		msg := diagnosticMessage(d)
		if strings.Contains(msg, "unrecognized character") {
			t.Fatalf("unexpected diagnostic: %v", msg)
		}
	}
}

func TestAnalyzeSyntaxError(t *testing.T) {
	src := `let x = ;`
	diags := Analyze(src)
	if len(diags) == 0 {
		t.Fatal("expected at least one diagnostic for syntax error")
	}
}

func TestAnalyzeMissingActive(t *testing.T) {
	src := `
options iteration=1 duration=10;
xiangling char lvl=90/90 cons=0 talent=1,1,1;
xiangling add weapon="thecatch" refine=1 lvl=90/90;
xiangling add set="emblemofseveredfate" count=4;
xiangling add stats hp=4780 atk=311;
target lvl=100 resist=0.1;
xiangling skill;
`
	diags := Analyze(src)
	found := false
	var msgs []string
	for _, d := range diags {
		msg := diagnosticMessage(d)
		msgs = append(msgs, msg)
		if strings.Contains(msg, "active") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected diagnostic about missing active char, got %v", msgs)
	}
}

func TestCompleteFiltersPrefix(t *testing.T) {
	items := mustCompletionItems(t, Complete("xia", 0, 3))
	if len(items) == 0 {
		t.Fatal("expected completions for prefix xia")
	}
	for _, it := range items {
		if !strings.HasPrefix(strings.ToLower(it.Label), "xia") {
			t.Fatalf("item %q does not match prefix xia", it.Label)
		}
	}
}

func TestHoverCharacter(t *testing.T) {
	text := "xiangling skill;"
	h := HoverAt(text, 0, 2)
	if h == nil {
		t.Fatal("expected hover for character name")
	}
}

func TestWordAt(t *testing.T) {
	text := "xiangling skill;"
	w, _, _ := WordAt(text, 0, 3)
	if w != "xiangling" {
		t.Fatalf("WordAt = %q, want xiangling", w)
	}
}

func TestDocumentStoreFullSync(t *testing.T) {
	s := NewStore()
	u := uri.File("/tmp/test.gcsim")
	s.Open(u, 1, "a")
	s.Set(u, 2, "abc")
	d, ok := s.Get(u)
	if !ok || d.Text != "abc" || d.Version != 2 {
		t.Fatalf("store state = %+v ok=%v", d, ok)
	}
	s.Close(u)
	if _, ok := s.Get(u); ok {
		t.Fatal("expected document closed")
	}
}

func diagnosticMessage(d protocol.Diagnostic) string {
	switch m := d.Message.(type) {
	case protocol.String:
		return string(m)
	case *protocol.MarkupContent:
		if m != nil {
			return m.Value
		}
	}
	return ""
}

func mustCompletionItems(t *testing.T, res protocol.CompletionResult) []protocol.CompletionItem {
	t.Helper()
	switch v := res.(type) {
	case protocol.CompletionItemSlice:
		return []protocol.CompletionItem(v)
	case *protocol.CompletionList:
		if v == nil {
			return nil
		}
		return v.Items
	default:
		t.Fatalf("unexpected completion result type %T", res)
		return nil
	}
}
