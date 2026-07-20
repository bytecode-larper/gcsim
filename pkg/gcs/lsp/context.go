package lsp

import (
	"strings"
	"unicode"
)

// completeScope restricts which catalog entries are offered.
type completeScope int

const (
	scopeAny completeScope = iota
	scopeWeapon
	scopeSet
	scopeCharacter
	scopeAction
	scopeStat
	scopeElement
	scopeOption
	scopeField
	scopeActionParam
)

// completionContext inspects text before the cursor and returns a scope plus
// the prefix that should filter labels (already lower-cased, no quotes).
//
// Line/token based (not a full re-parse) so incomplete buffers still work.
func completionContext(text string, line, character uint32) (scope completeScope, filter string) {
	offset := utf16Offset(text, line, character)
	if offset > len(text) {
		offset = len(text)
	}
	before := text[:offset]
	filter = prefixBefore(before)

	// weapon="…  / set="…  (cursor inside the string)
	if key, ok := stringAssignKey(before); ok {
		return scopeForAssignKey(key), filter
	}

	// weapon=wid  / set=obs  (no quotes, or quotes not typed yet)
	if key, ok := bareAssignKey(before); ok {
		if s := scopeForAssignKey(key); s != scopeAny {
			return s, filter
		}
	}

	// .mualani.nightsoul.|
	if isFieldCompletion(before) {
		return scopeField, filter
	}

	// walk[f=|  or +params=[stacks=|
	if isParamBracketContext(before) {
		return scopeActionParam, filter
	}

	lineStart := before
	if i := strings.LastIndexByte(before, '\n'); i >= 0 {
		lineStart = before[i+1:]
	}
	trimLeft := strings.TrimLeftFunc(lineStart, unicode.IsSpace)
	lowerLine := strings.ToLower(trimLeft)

	switch {
	case hasLinePrefix(lowerLine, "active"):
		return scopeCharacter, filter
	case hasLinePrefix(lowerLine, "options"):
		return scopeOption, filter
	case isStatsLine(lowerLine):
		return scopeStat, filter
	case isActionLineCompletion(lowerLine):
		return scopeAction, filter
	}

	return scopeAny, filter
}

func scopeForAssignKey(key string) completeScope {
	switch strings.ToLower(key) {
	case "weapon":
		return scopeWeapon
	case "set":
		return scopeSet
	case "element", "particle_element":
		return scopeElement
	default:
		return scopeAny
	}
}

// stringAssignKey: cursor is inside an unclosed "… after <key> = "
func stringAssignKey(before string) (key string, ok bool) {
	line := before
	if i := strings.LastIndexByte(before, '\n'); i >= 0 {
		line = before[i+1:]
	}
	if strings.Count(line, "\"")%2 == 0 {
		return "", false
	}
	q := strings.LastIndexByte(before, '"')
	if q < 0 {
		return "", false
	}
	i := q - 1
	for i >= 0 && (before[i] == ' ' || before[i] == '\t') {
		i--
	}
	if i < 0 || before[i] != '=' {
		return "", false
	}
	i--
	for i >= 0 && (before[i] == ' ' || before[i] == '\t') {
		i--
	}
	end := i + 1
	for i >= 0 && isIdentByte(before[i]) {
		i--
	}
	key = before[i+1 : end]
	if key == "" {
		return "", false
	}
	return key, true
}

// bareAssignKey: <key>=partial at end of before (not inside a string)
func bareAssignKey(before string) (key string, ok bool) {
	if _, in := stringAssignKey(before); in {
		return "", false
	}
	i := len(before) - 1
	for i >= 0 && isIdentByte(before[i]) {
		i--
	}
	for i >= 0 && (before[i] == ' ' || before[i] == '\t') {
		i--
	}
	if i < 0 || before[i] != '=' {
		return "", false
	}
	i--
	for i >= 0 && (before[i] == ' ' || before[i] == '\t') {
		i--
	}
	end := i + 1
	for i >= 0 && isIdentByte(before[i]) {
		i--
	}
	key = before[i+1 : end]
	if key == "" {
		return "", false
	}
	return key, true
}

func prefixBefore(before string) string {
	if len(before) == 0 {
		return ""
	}
	line := before[afterLastNewline(before):]
	if strings.Count(line, "\"")%2 == 1 {
		q := strings.LastIndexByte(before, '"')
		return strings.ToLower(before[q+1:])
	}
	i := len(before)
	for i > 0 && isIdentByte(before[i-1]) {
		i--
	}
	return strings.ToLower(before[i:])
}

func afterLastNewline(s string) int {
	if i := strings.LastIndexByte(s, '\n'); i >= 0 {
		return i + 1
	}
	return 0
}

func hasLinePrefix(lowerLine, keyword string) bool {
	if !strings.HasPrefix(lowerLine, keyword) {
		return false
	}
	rest := lowerLine[len(keyword):]
	return rest == "" || rest[0] == ' ' || rest[0] == '\t'
}

func isFieldCompletion(before string) bool {
	i := len(before)
	for i > 0 && isIdentByte(before[i-1]) {
		i--
	}
	return i > 0 && before[i-1] == '.'
}

func isParamBracketContext(before string) bool {
	line := before[afterLastNewline(before):]
	depth := 0
	for j := 0; j < len(line); j++ {
		switch line[j] {
		case '[':
			depth++
		case ']':
			if depth > 0 {
				depth--
			}
		}
	}
	return depth > 0
}

func isStatsLine(lowerLine string) bool {
	// "x add stats …" or "… stats hp=…"
	return strings.Contains(lowerLine, " stats") || strings.HasSuffix(lowerLine, "stats")
}

func isActionLineCompletion(lowerLine string) bool {
	fields := strings.Fields(lowerLine)
	if len(fields) < 2 {
		return false
	}
	switch fields[0] {
	case "options", "active", "target", "energy", "hurt", "fn", "let", "if",
		"while", "for", "switch", "return", "else", "case", "default":
		return false
	}
	switch fields[1] {
	case "char", "add":
		return false
	}
	return true
}

// detailMatchesScope uses CompletionItem.Detail strings from buildCatalog.
func detailMatchesScope(detail string, scope completeScope) bool {
	switch scope {
	case scopeAny:
		return true
	case scopeWeapon:
		return detail == "weapon"
	case scopeSet:
		return detail == "artifact set"
	case scopeCharacter:
		return detail == "character"
	case scopeAction:
		return detail == "action"
	case scopeStat:
		return detail == "stat"
	case scopeElement:
		return detail == "element"
	case scopeOption:
		return detail == "option"
	case scopeField:
		return detail == "field"
	case scopeActionParam:
		return detail == "action/weapon param"
	default:
		return true
	}
}