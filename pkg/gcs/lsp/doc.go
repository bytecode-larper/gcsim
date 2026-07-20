// Package lsp implements a Language Server Protocol (LSP) server for gcsim
// configuration scripts (gcsl).
//
// The server uses [go.lsp.dev/protocol] and reuses the existing gcsl lexer and
// parser in pkg/gcs for diagnostics, so editor feedback matches what the
// simulator accepts. It also indexes the document AST for go-to-definition,
// document symbols, and hover/completion of user-defined functions.
//
// Completion/hover names are derived from language tables (ast.Keywords,
// ast.ActionKeys, ast.StatKeys, shortcut maps, eval.BuiltinSysFuncs). Human
// prose lives only in docs.go overlays.
//
// The cmd/gcsls binary exposes this package over stdio.
package lsp
