package lsp

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

// Server implements a gcsl language server using go.lsp.dev/protocol.
type Server struct {
	protocol.UnimplementedServer

	mu     sync.Mutex
	client protocol.Client
	store  *Store
	log    *slog.Logger

	// rootURI is the optional workspace root from initialize.
	rootURI uri.URI
}

// NewServer constructs a language server. logger may be nil (defaults to stderr).
func NewServer(logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return &Server{
		store: NewStore(),
		log:   logger,
	}
}

// SetClient stores the client dispatcher returned by protocol.NewServer.
// Must be called before the connection begins handling requests that publish diagnostics.
func (s *Server) SetClient(c protocol.Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.client = c
}

func (s *Server) getClient() protocol.Client {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.client
}

// Initialize handles the LSP initialize request.
func (s *Server) Initialize(_ context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	if params != nil && params.RootURI != nil {
		s.rootURI = *params.RootURI
	}

	openClose := true
	change := protocol.TextDocumentSyncKindFull
	triggerChars := []string{".", "\""}

	return &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			PositionEncoding: protocol.PositionEncodingKindUTF16,
			TextDocumentSync: &protocol.TextDocumentSyncOptions{
				OpenClose: &openClose,
				Change:    &change,
			},
			CompletionProvider: &protocol.CompletionOptions{
				TriggerCharacters: triggerChars,
			},
			HoverProvider:          protocol.Boolean(true),
			DefinitionProvider:     protocol.Boolean(true),
			DocumentSymbolProvider: protocol.Boolean(true),
		},
		ServerInfo: protocol.ServerInfo{
			Name:    "gcsls",
			Version: protocol.NewOptional("0.2.0"),
		},
	}, nil
}

// Initialized handles the initialized notification.
func (s *Server) Initialized(_ context.Context, _ *protocol.InitializedParams) error {
	s.log.Info("gcsl language server initialized")
	return nil
}

// Shutdown handles shutdown.
func (s *Server) Shutdown(_ context.Context) error {
	s.log.Info("shutdown requested")
	return nil
}

// Exit handles exit (notification).
func (s *Server) Exit(_ context.Context) error {
	s.log.Info("exit")
	// Closing is handled by the process main after conn ends; force exit code 0 path.
	return nil
}

// DidOpen handles textDocument/didOpen.
func (s *Server) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	if params == nil {
		return nil
	}
	doc := params.TextDocument
	s.store.Open(doc.URI, doc.Version, doc.Text)
	s.publishDiagnostics(ctx, doc.URI, doc.Version, doc.Text)
	return nil
}

// DidChange handles textDocument/didChange (full document sync).
func (s *Server) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	if params == nil {
		return nil
	}
	docURI := params.TextDocument.URI
	version := params.TextDocument.Version

	// Prefer last whole-document change; fall back to applying partials in order.
	text := ""
	haveWhole := false
	for _, change := range params.ContentChanges {
		switch c := change.(type) {
		case *protocol.TextDocumentContentChangeWholeDocument:
			text = c.Text
			haveWhole = true
		case *protocol.TextDocumentContentChangePartial:
			if haveWhole {
				// apply partial on top of whole text we already took
				text = applyRangeEdit(text, c.Range.Start.Line, c.Range.Start.Character, c.Range.End.Line, c.Range.End.Character, c.Text)
			} else {
				s.store.ApplyChange(docURI, version,
					c.Range.Start.Line, c.Range.Start.Character,
					c.Range.End.Line, c.Range.End.Character,
					c.Text, false)
			}
		}
	}
	if haveWhole {
		s.store.Set(docURI, version, text)
	}

	d, ok := s.store.Get(docURI)
	if !ok {
		return nil
	}
	s.publishDiagnostics(ctx, docURI, d.Version, d.Text)
	return nil
}

// DidClose handles textDocument/didClose.
func (s *Server) DidClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	if params == nil {
		return nil
	}
	docURI := params.TextDocument.URI
	s.store.Close(docURI)
	// Clear diagnostics for closed file.
	client := s.getClient()
	if client != nil {
		_ = client.PublishDiagnostics(ctx, &protocol.PublishDiagnosticsParams{
			URI:         docURI,
			Diagnostics: []protocol.Diagnostic{},
		})
	}
	return nil
}

// Completion handles textDocument/completion.
func (s *Server) Completion(_ context.Context, params *protocol.CompletionParams) (protocol.CompletionResult, error) {
	if params == nil {
		return protocol.CompletionItemSlice{}, nil
	}
	idx, text, ok := s.store.IndexOf(params.TextDocument.URI)
	if !ok {
		return protocol.CompletionItemSlice{}, nil
	}
	return CompleteWithIndex(text, idx, params.Position.Line, params.Position.Character), nil
}

// Hover handles textDocument/hover.
func (s *Server) Hover(_ context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	if params == nil {
		return nil, nil
	}
	idx, text, ok := s.store.IndexOf(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}
	return HoverAtWithIndex(text, idx, params.Position.Line, params.Position.Character), nil
}

// Definition handles textDocument/definition (go-to-definition).
func (s *Server) Definition(_ context.Context, params *protocol.DefinitionParams) (protocol.DefinitionResult, error) {
	if params == nil {
		return nil, nil
	}
	idx, text, ok := s.store.IndexOf(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}
	return DefinitionAtWithIndex(params.TextDocument.URI, text, idx, params.Position.Line, params.Position.Character), nil
}

// DocumentSymbol handles textDocument/documentSymbol (outline).
func (s *Server) DocumentSymbol(_ context.Context, params *protocol.DocumentSymbolParams) (protocol.DocumentSymbolResult, error) {
	if params == nil {
		return protocol.DocumentSymbolSlice{}, nil
	}
	idx, text, ok := s.store.IndexOf(params.TextDocument.URI)
	if !ok {
		return protocol.DocumentSymbolSlice{}, nil
	}
	return DocumentSymbolsFromIndex(text, idx), nil
}

func (s *Server) publishDiagnostics(ctx context.Context, docURI uri.URI, version int32, text string) {
	client := s.getClient()
	if client == nil {
		return
	}
	diags := Analyze(text)
	params := &protocol.PublishDiagnosticsParams{
		URI:         docURI,
		Diagnostics: diags,
	}
	if version != 0 {
		params.Version = protocol.NewOptional(version)
	}
	if err := client.PublishDiagnostics(ctx, params); err != nil {
		s.log.Warn("publishDiagnostics failed", "err", err)
	} else {
		s.log.Debug("published diagnostics", "uri", docURI, "count", len(diags))
	}
}

// RunStdio serves the language server over stdin/stdout until the connection ends.
func RunStdio(ctx context.Context, logger *slog.Logger) error {
	srv := NewServer(logger)
	stream := jsonrpc2.NewStream(stdrwc{})
	ctx, conn, client := protocol.NewServer(ctx, srv, stream)
	srv.SetClient(client)

	logger.Info("gcsls listening on stdio")
	<-conn.Done()
	if err := conn.Err(); err != nil {
		return err
	}
	return nil
}

// stdrwc adapts os.Stdin/Stdout to io.ReadWriteCloser for jsonrpc2.
type stdrwc struct{}

func (stdrwc) Read(p []byte) (int, error)  { return os.Stdin.Read(p) }
func (stdrwc) Write(p []byte) (int, error) { return os.Stdout.Write(p) }
func (stdrwc) Close() error {
	if err := os.Stdin.Close(); err != nil {
		return err
	}
	return os.Stdout.Close()
}
