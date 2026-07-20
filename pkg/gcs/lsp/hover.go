package lsp

import (
	"fmt"
	"strings"

	"go.lsp.dev/protocol"
)

// HoverAt returns hover information for the identifier at the given position.
func HoverAt(text string, line, character uint32) *protocol.Hover {
	return HoverAtWithIndex(text, IndexDocument(text), line, character)
}

// HoverAtWithIndex is like HoverAt but reuses a prebuilt symbol index.
func HoverAtWithIndex(text string, idx *Index, line, character uint32) *protocol.Hover {
	word, startByte, endByte := WordAt(text, line, character)
	rawWord := word
	word = strings.TrimLeft(word, ".")
	if word == "" {
		return nil
	}

	// Prefer the last segment of a field path (.xiangling.energy → energy / xiangling)
	parts := strings.Split(word, ".")
	cleanParts := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			cleanParts = append(cleanParts, p)
		}
	}
	if len(cleanParts) == 0 {
		return nil
	}

	// Full identifier under cursor (no field split) for local symbols / calls.
	primary := cleanParts[len(cleanParts)-1]
	if !strings.Contains(rawWord, ".") {
		primary = word
	}

	assignKey := looksLikeAssignKey(text, endByte)

	// 1) Document-local symbols (user functions, lets, params)
	if idx != nil {
		if sym := idx.Lookup(primary); sym != nil {
			return hoverMarkdown(describeSymbol(sym), text, startByte, endByte)
		}
	}

	// 2) Catalog — try primary, then other path segments
	candidates := []string{primary}
	if len(cleanParts) > 0 {
		candidates = append(candidates, cleanParts[0])
		if word != primary {
			candidates = append(candidates, word)
		}
	}
	seen := map[string]bool{}
	for _, c := range candidates {
		c = strings.TrimLeft(c, ".")
		if c == "" || seen[c] {
			continue
		}
		seen[c] = true
		if md := describeCatalog(c, assignKey); md != "" {
			return hoverMarkdown(md, text, startByte, endByte)
		}
	}

	return nil
}

func describeSymbol(s *Symbol) string {
	switch s.Kind {
	case SymbolFunction:
		return fmt.Sprintf("**`%s`** — user function\n\n```gcsl\n%s\n```", s.Name, s.Detail)
	case SymbolParameter:
		return fmt.Sprintf("**`%s`** — %s", s.Name, s.Detail)
	case SymbolVariable:
		return fmt.Sprintf("**`%s`** — local variable", s.Name)
	default:
		return fmt.Sprintf("**`%s`**", s.Name)
	}
}

func hoverMarkdown(md, text string, startByte, endByte int) *protocol.Hover {
	startLine, startChar := LineColFromByte(text, startByte)
	endLine, endChar := LineColFromByte(text, endByte)
	return &protocol.Hover{
		Contents: &protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: md,
		},
		Range: &protocol.Range{
			Start: protocol.Position{Line: startLine, Character: startChar},
			End:   protocol.Position{Line: endLine, Character: endChar},
		},
	}
}

// looksLikeAssignKey reports whether the word ending at endByte is followed by '='
// (e.g. hold=1, f=1, interval=480). Used to prefer param/config-key docs over
// system functions that share names (notably `f`).
func looksLikeAssignKey(text string, endByte int) bool {
	if endByte < 0 {
		endByte = 0
	}
	i := endByte
	for i < len(text) {
		switch text[i] {
		case ' ', '\t':
			i++
			continue
		case '=':
			return true
		default:
			return false
		}
	}
	return false
}

// describe is kept for tests / simple catalog-only lookup without position context.
func describe(word string) string {
	return describeCatalog(word, false)
}
