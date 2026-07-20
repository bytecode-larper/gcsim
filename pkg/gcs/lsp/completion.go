package lsp

import (
	"strings"

	"go.lsp.dev/protocol"
)

// Complete returns completion items for the given document position.
// Includes the global catalog plus user-defined functions from the document,
// filtered by syntactic context (e.g. only weapons inside weapon="…").
func Complete(text string, line, character uint32) protocol.CompletionResult {
	return CompleteWithIndex(text, IndexDocument(text), line, character)
}

// CompleteWithIndex is like Complete but reuses a prebuilt symbol index.
func CompleteWithIndex(text string, idx *Index, line, character uint32) protocol.CompletionResult {
	scope, filter := completionContext(text, line, character)

	cat := getCatalog()
	out := make(protocol.CompletionItemSlice, 0, 32)
	seen := make(map[string]struct{})

	add := func(item protocol.CompletionItem, detail string) {
		key := strings.ToLower(item.Label)
		if _, ok := seen[key]; ok {
			return
		}
		if filter != "" && !strings.HasPrefix(key, filter) {
			return
		}
		if !detailMatchesScope(detail, scope) {
			return
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}

	// User functions only in unrestricted contexts (calls / general editing).
	if scope == scopeAny && idx != nil {
		for _, s := range idx.FunctionList {
			add(protocol.CompletionItem{
				Label:  s.Name,
				Kind:   protocol.CompletionItemKindFunction,
				Detail: protocol.NewOptional(s.Detail),
			}, "user function")
		}
	}

	for _, item := range cat.items {
		detail, _ := item.Detail.Get()
		add(item, detail)
	}
	return out
}
