package lsp

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/genshinsim/gcsim/pkg/gcs/ast"
	"github.com/genshinsim/gcsim/pkg/gcs/parser"
)

// SymbolKind classifies document-local symbols.
type SymbolKind int

const (
	SymbolFunction SymbolKind = iota
	SymbolVariable
	SymbolParameter
)

// Symbol is a named declaration in a document.
type Symbol struct {
	Name      string
	Kind      SymbolKind
	// StartByte/EndByte cover the identifier token.
	StartByte int
	EndByte   int
	// Detail is a short signature / type hint for hover and outline.
	Detail string
	// Args is set for functions (parameter names).
	Args []string
}

// Index holds declarations found in a single document.
type Index struct {
	// Functions keyed by lower-case name (last declaration wins).
	Functions map[string]*Symbol
	// Variables keyed by lower-case name. Includes function params and lets.
	// Last declaration wins for simple lookup (good enough for configs).
	Variables map[string]*Symbol
	// Ordered function list for outline (source order).
	FunctionList []*Symbol
}

// Lookup finds a function or variable by name (case-insensitive).
// Functions are preferred when both exist (unusual in gcsl).
func (idx *Index) Lookup(name string) *Symbol {
	if idx == nil {
		return nil
	}
	key := strings.ToLower(name)
	if s, ok := idx.Functions[key]; ok {
		return s
	}
	if s, ok := idx.Variables[key]; ok {
		return s
	}
	return nil
}

// LookupFunction finds only user functions.
func (idx *Index) LookupFunction(name string) *Symbol {
	if idx == nil {
		return nil
	}
	return idx.Functions[strings.ToLower(name)]
}

// IndexDocument parses src and collects declarations.
// On hard parse failure, falls back to a line scan for `fn` / `let` so
// go-to-definition still works while editing incomplete buffers.
func IndexDocument(src string) *Index {
	idx := &Index{
		Functions: make(map[string]*Symbol),
		Variables: make(map[string]*Symbol),
	}

	file := ast.NewFile()
	p := parser.New(file, src)
	_, prog, err := p.Parse()
	if err == nil && prog != nil {
		collectNode(src, prog, idx)
		return idx
	}

	// Fallback: regex scan for top-level-ish declarations.
	indexByScan(src, idx)
	return idx
}

func collectNode(src string, node ast.Node, idx *Index) {
	if node == nil {
		return
	}
	switch n := node.(type) {
	case *ast.BlockStmt:
		for _, c := range n.List {
			collectNode(src, c, idx)
		}
	case *ast.FnStmt:
		addFnSymbol(src, n, idx)
		if n.Func != nil {
			for _, a := range n.Func.Args {
				if a == nil {
					continue
				}
				start := int(a.Pos)
				end := start + len(a.Value)
				s := &Symbol{
					Name:      a.Value,
					Kind:      SymbolParameter,
					StartByte: start,
					EndByte:   end,
					Detail:    fmt.Sprintf("parameter of %s", n.Ident.Val),
				}
				idx.Variables[strings.ToLower(a.Value)] = s
			}
			if n.Func.Body != nil {
				collectNode(src, n.Func.Body, idx)
			}
		}
	case *ast.LetStmt:
		addLetSymbol(src, n, idx)
	case *ast.IfStmt:
		if n.IfBlock != nil {
			collectNode(src, n.IfBlock, idx)
		}
		if n.ElseBlock != nil {
			collectNode(src, n.ElseBlock, idx)
		}
	case *ast.WhileStmt:
		if n.WhileBlock != nil {
			collectNode(src, n.WhileBlock, idx)
		}
	case *ast.ForStmt:
		if n.Init != nil {
			collectNode(src, n.Init, idx)
		}
		if n.Body != nil {
			collectNode(src, n.Body, idx)
		}
	case *ast.SwitchStmt:
		for _, c := range n.Cases {
			if c != nil && c.Body != nil {
				collectNode(src, c.Body, idx)
			}
		}
		if n.Default != nil {
			collectNode(src, n.Default, idx)
		}
	case *ast.CaseStmt:
		if n.Body != nil {
			collectNode(src, n.Body, idx)
		}
	}
}

func addFnSymbol(src string, n *ast.FnStmt, idx *Index) {
	name := n.Ident.Val
	start := int(n.Ident.Pos)
	end := start + len(name)
	// Guard against bogus positions
	if start < 0 || start > len(src) {
		start = 0
		end = 0
	}
	if end > len(src) {
		end = len(src)
	}

	var args []string
	if n.Func != nil {
		for _, a := range n.Func.Args {
			if a != nil {
				args = append(args, a.Value)
			}
		}
	}
	detail := fmt.Sprintf("fn %s(%s)", name, strings.Join(args, ", "))
	s := &Symbol{
		Name:      name,
		Kind:      SymbolFunction,
		StartByte: start,
		EndByte:   end,
		Detail:    detail,
		Args:      args,
	}
	key := strings.ToLower(name)
	idx.Functions[key] = s
	idx.FunctionList = append(idx.FunctionList, s)
}

func addLetSymbol(src string, n *ast.LetStmt, idx *Index) {
	name := n.Ident.Val
	start := int(n.Ident.Pos)
	end := start + len(name)
	if start < 0 || start > len(src) {
		start = 0
		end = 0
	}
	if end > len(src) {
		end = len(src)
	}
	s := &Symbol{
		Name:      name,
		Kind:      SymbolVariable,
		StartByte: start,
		EndByte:   end,
		Detail:    "local variable",
	}
	idx.Variables[strings.ToLower(name)] = s
}

var (
	reFnDecl  = regexp.MustCompile(`(?m)(?:^|[\s;{}])fn\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(`)
	reLetDecl = regexp.MustCompile(`(?m)(?:^|[\s;{}])let\s+([A-Za-z_][A-Za-z0-9_]*)\b`)
)

func indexByScan(src string, idx *Index) {
	for _, m := range reFnDecl.FindAllStringSubmatchIndex(src, -1) {
		// submatch 1 is the name
		if len(m) < 4 {
			continue
		}
		start, end := m[2], m[3]
		name := src[start:end]
		s := &Symbol{
			Name:      name,
			Kind:      SymbolFunction,
			StartByte: start,
			EndByte:   end,
			Detail:    fmt.Sprintf("fn %s(...)", name),
		}
		key := strings.ToLower(name)
		idx.Functions[key] = s
		idx.FunctionList = append(idx.FunctionList, s)
	}
	for _, m := range reLetDecl.FindAllStringSubmatchIndex(src, -1) {
		if len(m) < 4 {
			continue
		}
		start, end := m[2], m[3]
		name := src[start:end]
		s := &Symbol{
			Name:      name,
			Kind:      SymbolVariable,
			StartByte: start,
			EndByte:   end,
			Detail:    "local variable",
		}
		idx.Variables[strings.ToLower(name)] = s
	}
}

// symbolRange converts a symbol's byte span to an LSP range.
func symbolRange(text string, s *Symbol) protocolRange {
	sl, sc := LineColFromByte(text, s.StartByte)
	el, ec := LineColFromByte(text, s.EndByte)
	return protocolRange{
		startLine: sl, startChar: sc,
		endLine: el, endChar: ec,
	}
}

// small helper to avoid importing protocol in pure logic paths if needed
type protocolRange struct {
	startLine, startChar, endLine, endChar uint32
}
