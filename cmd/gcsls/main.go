// Command gcsls is a Language Server Protocol implementation for gcsim config
// files (gcsl). It speaks LSP over stdio using go.lsp.dev/protocol.
//
// Features:
//   - Diagnostics from the real gcsl parser (syntax + config validation)
//   - Completions for keywords, characters, actions, weapons, sets, stats, system functions
//   - Hover documentation for the same identifiers
//
// Example VS Code / Neovim client settings: start this binary with no args and
// communicate over stdin/stdout. Associate filetypes "gcsim", "gcsl", or plain
// text configs as desired.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/genshinsim/gcsim/pkg/gcs/lsp"
)

func main() {
	// All logs must go to stderr — stdout is the LSP transport.
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := lsp.RunStdio(ctx, logger); err != nil {
		logger.Error("language server stopped with error", "err", err)
		os.Exit(1)
	}
}
