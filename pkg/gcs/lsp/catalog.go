package lsp

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/genshinsim/gcsim/pkg/gcs/ast"
	"github.com/genshinsim/gcsim/pkg/gcs/eval"
	"github.com/genshinsim/gcsim/pkg/shortcut"
	"go.lsp.dev/protocol"
)

// Completion catalog built once from language tables + doc overlays.
type catalog struct {
	items []protocol.CompletionItem
}

var (
	catalogOnce sync.Once
	globalCat   *catalog
)

func getCatalog() *catalog {
	catalogOnce.Do(func() {
		globalCat = buildCatalog()
	})
	return globalCat
}

func buildCatalog() *catalog {
	var items []protocol.CompletionItem
	seen := make(map[string]struct{})

	add := func(label, detail string, kind protocol.CompletionItemKind) {
		if _, ok := seen[label]; ok {
			return
		}
		seen[label] = struct{}{}
		items = append(items, protocol.CompletionItem{
			Label:  label,
			Kind:   kind,
			Detail: protocol.NewOptional(detail),
		})
	}

	// Names from real language tables
	for _, word := range ast.Keywords() {
		add(word, "keyword", protocol.CompletionItemKindKeyword)
	}
	for word := range ast.ActionKeys {
		add(word, "action", protocol.CompletionItemKindFunction)
	}
	for word := range ast.StatKeys {
		add(word, "stat", protocol.CompletionItemKindProperty)
	}
	for word := range ast.EleKeys {
		add(word, "element", protocol.CompletionItemKindEnumMember)
	}
	for word := range shortcut.CharNameToKey {
		add(word, "character", protocol.CompletionItemKindClass)
	}
	for word := range shortcut.WeaponNameToKey {
		add(word, "weapon", protocol.CompletionItemKindValue)
	}
	for word := range shortcut.SetNameToKey {
		add(word, "artifact set", protocol.CompletionItemKindValue)
	}
	for _, word := range eval.BuiltinSysFuncs {
		add(word, "system function", protocol.CompletionItemKindFunction)
	}

	// Names only documented in LSP overlays (no shared runtime table yet)
	for word := range optionDocs {
		add(word, "option", protocol.CompletionItemKindProperty)
	}
	for word := range fieldDocs {
		add(word, "field", protocol.CompletionItemKindField)
	}
	for word := range configKeyDocs {
		add(word, "config key", protocol.CompletionItemKindProperty)
	}
	for word := range actionParamDocs {
		add(word, "action/weapon param", protocol.CompletionItemKindProperty)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Label < items[j].Label
	})

	return &catalog{items: items}
}

// describeCatalog returns markdown for a global catalog name, or "".
// preferAssignKey biases toward config/action param docs when the word is an assignment key.
func describeCatalog(word string, preferAssignKey bool) string {
	lw := strings.ToLower(word)

	tryAssign := func() string {
		if d, ok := actionParamDocs[lw]; ok {
			return fmt.Sprintf("**`%s`** — action/weapon parameter\n\n%s", lw, d)
		}
		if d, ok := configKeyDocs[lw]; ok {
			return fmt.Sprintf("**`%s`** — config key\n\n%s", lw, d)
		}
		if _, ok := ast.StatKeys[lw]; ok {
			return fmt.Sprintf("**`%s`** — artifact/character stat key", lw)
		}
		if d, ok := optionDocs[lw]; ok {
			return fmt.Sprintf("**`%s`** — simulator option\n\n%s\n\nUsed in `options %s=...;`", lw, d, lw)
		}
		return ""
	}

	if preferAssignKey {
		if h := tryAssign(); h != "" {
			return h
		}
	}

	if ast.IsKeyword(lw) {
		d := docOr(keywordDocs, lw, "")
		if d != "" {
			return fmt.Sprintf("**`%s`** — keyword\n\n%s", lw, d)
		}
		return fmt.Sprintf("**`%s`** — keyword", lw)
	}
	if _, ok := ast.ActionKeys[lw]; ok {
		d := docOr(actionDocs, lw, "Character action")
		return fmt.Sprintf("**`%s`** — character action\n\n%s\n\nUsage: `<char> %s;`", lw, d, lw)
	}
	if _, ok := ast.StatKeys[lw]; ok {
		return fmt.Sprintf("**`%s`** — artifact/character stat key", lw)
	}
	if _, ok := ast.EleKeys[lw]; ok {
		return fmt.Sprintf("**`%s`** — element", lw)
	}
	if key, ok := shortcut.CharNameToKey[lw]; ok {
		return fmt.Sprintf("**`%s`** — character\n\nKey: `%s`", lw, key.String())
	}
	if key, ok := shortcut.WeaponNameToKey[lw]; ok {
		return fmt.Sprintf("**`%s`** — weapon\n\nKey: `%s`\n\nTypically: `weapon=\"%s\"`.", lw, key.String(), lw)
	}
	if key, ok := shortcut.SetNameToKey[lw]; ok {
		return fmt.Sprintf("**`%s`** — artifact set\n\nKey: `%s`\n\nTypically: `set=\"%s\"`.", lw, key.String(), lw)
	}
	if isBuiltinSysFunc(lw) {
		d := docOr(sysFuncDocs, lw, "System function")
		return fmt.Sprintf("**`%s`** — system function\n\n%s", lw, d)
	}
	if d, ok := optionDocs[lw]; ok {
		return fmt.Sprintf("**`%s`** — simulator option\n\n%s\n\nUsed in `options %s=...;`", lw, d, lw)
	}
	if d, ok := fieldDocs[lw]; ok {
		return fmt.Sprintf("**`%s`** — field segment\n\n%s\n\nFields start with `.` (e.g. `.mualani.nightsoul.state`).", lw, d)
	}
	if !preferAssignKey {
		if h := tryAssign(); h != "" {
			return h
		}
	}
	return ""
}

func isBuiltinSysFunc(name string) bool {
	for _, s := range eval.BuiltinSysFuncs {
		if s == name {
			return true
		}
	}
	return false
}

// helpers used by tests
func keywordList() map[string]struct{} {
	m := make(map[string]struct{})
	for _, k := range ast.Keywords() {
		m[k] = struct{}{}
	}
	return m
}

func actionList() map[string]struct{} {
	m := make(map[string]struct{}, len(ast.ActionKeys))
	for k := range ast.ActionKeys {
		m[k] = struct{}{}
	}
	return m
}

func sysFuncs() []string {
	out := append([]string(nil), eval.BuiltinSysFuncs...)
	sort.Strings(out)
	return out
}

func optionKeys() []string {
	return sortedMapKeys(optionDocs)
}

func commonFieldSegments() []string {
	return sortedMapKeys(fieldDocs)
}

func sortedMapKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
