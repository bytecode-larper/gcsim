package lsp

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/genshinsim/gcsim/pkg/gcs/ast"
	"github.com/genshinsim/gcsim/pkg/gcs/parser"
	"go.lsp.dev/protocol"
)

var lnPrefix = regexp.MustCompile(`(?i)^ln\s*(\d+)(?::(\d+))?\s*:?\s*`)

// Analyze parses src with the real gcsl parser and returns LSP diagnostics.
func Analyze(src string) []protocol.Diagnostic {
	file := ast.NewFile()
	p := parser.New(file, src)
	simcfg, _, err := p.Parse()

	var diags []protocol.Diagnostic
	if err != nil {
		diags = append(diags, diagnosticFromError(src, err))
		// Still try to surface nothing else — parse stopped early.
		return diags
	}

	if simcfg != nil {
		for _, e := range simcfg.Errors {
			diags = append(diags, diagnosticFromError(src, e))
		}
	}
	return diags
}

func diagnosticFromError(src string, err error) protocol.Diagnostic {
	msg := err.Error()
	line, col, ok := positionFromASTError(err)
	if !ok {
		line, col, msg = parseLineFromMessage(msg)
	}

	// Convert 1-based parser positions to 0-based LSP positions.
	var startLine, startChar uint32
	if line > 0 {
		startLine = uint32(line - 1)
	}
	if col > 0 {
		startChar = uint32(col - 1)
	}

	endChar := startChar + 1
	// Prefer spanning the rest of the line for visibility
	if lines := strings.Split(src, "\n"); int(startLine) < len(lines) {
		lineText := strings.TrimRight(lines[startLine], "\r")
		lineUnits := utf16Len(lineText)
		if startChar > lineUnits {
			startChar = lineUnits
		}
		if endChar < lineUnits {
			endChar = lineUnits
		}
		if endChar == startChar {
			endChar = startChar + 1
		}
	}

	return protocol.Diagnostic{
		Range: protocol.Range{
			Start: protocol.Position{Line: startLine, Character: startChar},
			End:   protocol.Position{Line: startLine, Character: endChar},
		},
		Severity: protocol.DiagnosticSeverityError,
		Source:   protocol.NewOptional("gcsl"),
		Message:  protocol.String(msg),
	}
}

func positionFromASTError(err error) (line, col int, ok bool) {
	var ae ast.Error
	if errors.As(err, &ae) && ae.Pos.IsValid() {
		return ae.Pos.Line, ae.Pos.Column, true
	}
	// pointer form
	var aep *ast.Error
	if errors.As(err, &aep) && aep != nil && aep.Pos.IsValid() {
		return aep.Pos.Line, aep.Pos.Column, true
	}
	return 0, 0, false
}

// parseLineFromMessage extracts "lnN:..." or "lnN:M:..." prefixes used by the
// parser/error formatting and returns a cleaned message.
func parseLineFromMessage(msg string) (line, col int, cleaned string) {
	if m := lnPrefix.FindStringSubmatch(msg); m != nil {
		line, _ = strconv.Atoi(m[1])
		if m[2] != "" {
			col, _ = strconv.Atoi(m[2])
		}
		cleaned = strings.TrimSpace(msg[len(m[0]):])
		if cleaned == "" {
			cleaned = msg
		}
		return line, col, cleaned
	}
	// soft errors often have no location
	return 1, 1, msg
}

// FormatDiagCount is a small helper for logging.
func FormatDiagCount(n int) string {
	return fmt.Sprintf("%d diagnostic(s)", n)
}
