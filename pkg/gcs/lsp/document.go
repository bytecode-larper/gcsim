package lsp

import (
	"strings"
	"sync"

	"go.lsp.dev/uri"
)

// Document is an open text document tracked by the language server.
type Document struct {
	URI     uri.URI
	Version int32
	Text    string

	// cached index for the current Text (invalidated on Set/ApplyChange)
	index *Index
}

// Store holds the set of open documents.
type Store struct {
	mu   sync.RWMutex
	docs map[uri.URI]*Document
}

// NewStore creates an empty document store.
func NewStore() *Store {
	return &Store{docs: make(map[uri.URI]*Document)}
}

// Open records a newly opened document.
func (s *Store) Open(uri uri.URI, version int32, text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.docs[uri] = &Document{URI: uri, Version: version, Text: text, index: nil}
}

// Set replaces document text and version (full sync).
func (s *Store) Set(uri uri.URI, version int32, text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if d, ok := s.docs[uri]; ok {
		d.Version = version
		d.Text = text
		d.index = nil
		return
	}
	s.docs[uri] = &Document{URI: uri, Version: version, Text: text}
}

// ApplyChange applies a single textDocument/didChange content change.
// Supports whole-document and ranged (UTF-16 code unit) updates.
func (s *Store) ApplyChange(docURI uri.URI, version int32, startLine, startChar, endLine, endChar uint32, newText string, whole bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.docs[docURI]
	if !ok {
		d = &Document{URI: docURI}
		s.docs[docURI] = d
	}
	if whole {
		d.Text = newText
	} else {
		d.Text = applyRangeEdit(d.Text, startLine, startChar, endLine, endChar, newText)
	}
	d.Version = version
	d.index = nil
}

// IndexOf returns a cached symbol index for the document, building it if needed.
func (s *Store) IndexOf(docURI uri.URI) (*Index, string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.docs[docURI]
	if !ok {
		return nil, "", false
	}
	if d.index == nil {
		d.index = IndexDocument(d.Text)
	}
	return d.index, d.Text, true
}

// Close removes a document from the store.
func (s *Store) Close(uri uri.URI) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.docs, uri)
}

// Get returns a copy of the document if present.
func (s *Store) Get(uri uri.URI) (*Document, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.docs[uri]
	if !ok {
		return nil, false
	}
	cp := *d
	return &cp, true
}

// applyRangeEdit replaces text in [start, end) using UTF-16 code unit offsets
// for character positions (LSP default encoding).
func applyRangeEdit(text string, startLine, startChar, endLine, endChar uint32, newText string) string {
	start := utf16Offset(text, startLine, startChar)
	end := utf16Offset(text, endLine, endChar)
	if start > len(text) {
		start = len(text)
	}
	if end < start {
		end = start
	}
	if end > len(text) {
		end = len(text)
	}
	return text[:start] + newText + text[end:]
}

// utf16Offset maps an LSP (line, character) position to a byte offset in text.
// character is measured in UTF-16 code units.
func utf16Offset(text string, line, character uint32) int {
	lines := splitLinesKeepEnds(text)
	if int(line) >= len(lines) {
		return len(text)
	}
	// sum full lines before target
	offset := 0
	for i := uint32(0); i < line; i++ {
		offset += len(lines[i])
	}
	lineText := lines[line]
	// strip trailing newline for character counting within the line content
	content := strings.TrimRight(lineText, "\r\n")
	byteInLine := utf16CharToByte(content, character)
	return offset + byteInLine
}

func splitLinesKeepEnds(text string) []string {
	if text == "" {
		return []string{""}
	}
	var lines []string
	start := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			lines = append(lines, text[start:i+1])
			start = i + 1
		}
	}
	if start <= len(text) {
		// remaining (may be empty after final \n — still one more empty line if ends with \n? LSP: last line without trailing newline)
		if start < len(text) || (start == len(text) && len(text) > 0 && text[len(text)-1] != '\n') {
			lines = append(lines, text[start:])
		} else if start == len(text) && (len(text) == 0 || text[len(text)-1] == '\n') {
			// file ends with newline: editor may still have a last empty line
			lines = append(lines, "")
		}
	}
	if len(lines) == 0 {
		lines = []string{""}
	}
	return lines
}

// utf16CharToByte converts a UTF-16 character offset within s to a byte offset.
func utf16CharToByte(s string, character uint32) int {
	var units uint32
	for i, r := range s {
		if units >= character {
			return i
		}
		if r <= 0xFFFF {
			units++
		} else {
			// surrogate pair
			units += 2
		}
		if units > character {
			// landed in the middle of a surrogate; clamp to this rune start
			return i
		}
	}
	return len(s)
}

// WordAt returns the identifier-like word at the given position and its
// byte range within the full text. Positions are 0-based line/UTF-16 char.
func WordAt(text string, line, character uint32) (word string, startByte, endByte int) {
	offset := utf16Offset(text, line, character)
	if offset > len(text) {
		offset = len(text)
	}
	// If cursor is at end of word, still select that word
	start := offset
	for start > 0 && isIdentByte(text[start-1]) {
		start--
	}
	end := offset
	for end < len(text) && isIdentByte(text[end]) {
		end++
	}
	// Fields start with '.'; include leading dots when present
	for start > 0 && text[start-1] == '.' {
		start--
	}
	return text[start:end], start, end
}

func isIdentByte(b byte) bool {
	return b == '_' || b == '-' || b == '%' ||
		(b >= 'a' && b <= 'z') ||
		(b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9')
}

// LineColFromByte returns 0-based line and UTF-16 character for a byte offset.
func LineColFromByte(text string, byteOffset int) (line, character uint32) {
	if byteOffset < 0 {
		byteOffset = 0
	}
	if byteOffset > len(text) {
		byteOffset = len(text)
	}
	line = 0
	lineStart := 0
	for i := 0; i < byteOffset; i++ {
		if text[i] == '\n' {
			line++
			lineStart = i + 1
		}
	}
	character = utf16Len(text[lineStart:byteOffset])
	return line, character
}

func utf16Len(s string) uint32 {
	var n uint32
	for _, r := range s {
		if r <= 0xFFFF {
			n++
		} else {
			n += 2
		}
	}
	return n
}
