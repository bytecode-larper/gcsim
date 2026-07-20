package lsp

import (
	"strings"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

// DefinitionAt returns the declaration location for the identifier at position.
// Supports user functions and let/param bindings from the document index.
func DefinitionAt(docURI uri.URI, text string, line, character uint32) protocol.DefinitionResult {
	return DefinitionAtWithIndex(docURI, text, IndexDocument(text), line, character)
}

// DefinitionAtWithIndex is like DefinitionAt but reuses a prebuilt symbol index.
func DefinitionAtWithIndex(docURI uri.URI, text string, idx *Index, line, character uint32) protocol.DefinitionResult {
	word, _, _ := WordAt(text, line, character)
	word = strings.TrimLeft(word, ".")
	if word == "" {
		return nil
	}
	// Field paths: take the last segment only for local symbols (rare).
	if i := strings.LastIndex(word, "."); i >= 0 {
		word = word[i+1:]
	}
	if word == "" || idx == nil {
		return nil
	}

	sym := idx.Lookup(word)
	if sym == nil {
		return nil
	}

	r := symbolRange(text, sym)
	loc := protocol.Location{
		URI: docURI,
		Range: protocol.Range{
			Start: protocol.Position{Line: r.startLine, Character: r.startChar},
			End:   protocol.Position{Line: r.endLine, Character: r.endChar},
		},
	}
	return &loc
}

// DocumentSymbols returns outline symbols (functions) for the document.
func DocumentSymbols(text string) protocol.DocumentSymbolResult {
	return DocumentSymbolsFromIndex(text, IndexDocument(text))
}

// DocumentSymbolsFromIndex is like DocumentSymbols but reuses a prebuilt index.
func DocumentSymbolsFromIndex(text string, idx *Index) protocol.DocumentSymbolResult {
	if idx == nil {
		return protocol.DocumentSymbolSlice{}
	}
	out := make(protocol.DocumentSymbolSlice, 0, len(idx.FunctionList))
	for _, s := range idx.FunctionList {
		s := s
		r := symbolRange(text, s)
		rng := protocol.Range{
			Start: protocol.Position{Line: r.startLine, Character: r.startChar},
			End:   protocol.Position{Line: r.endLine, Character: r.endChar},
		}
		detail := s.Detail
		out = append(out, protocol.DocumentSymbol{
			Name:           s.Name,
			Detail:         &detail,
			Kind:           protocol.SymbolKindFunction,
			Range:          rng,
			SelectionRange: rng,
		})
	}
	return out
}
