# gcsls

Language Server Protocol (LSP) implementation for **gcsim config scripts** (gcsl).

Built on [`go.lsp.dev/protocol`](https://github.com/go-language-server/protocol) and the real gcsl parser in `pkg/gcs`.

## Features

| Capability | Status |
| --- | --- |
| Diagnostics | Parser + config validation errors (`active`, targets, etc.) |
| Completion | Keywords, characters, actions, weapons, sets, stats, elements, system functions, option/config keys, action params, **user `fn`s** — **context-filtered** (e.g. only weapons inside `weapon="…"`) |
| Hover | Catalog docs + **user functions / locals / params**; assignment keys (e.g. `f=1`) prefer param docs over system `f()` |
| Go to definition | User `fn` and `let` / param bindings in the current document |
| Document symbols | Outline of user functions |
| Full document sync | Yes (`TextDocumentSyncKind.Full`) |

## Build

From the repo root:

```bash
go build -o gcsls ./cmd/gcsls
# or install on PATH for editor extensions:
go build -o ~/.local/bin/gcsls ./cmd/gcsls
```

## Run

The server speaks LSP over **stdio** (stdout is the protocol stream; logs go to stderr):

```bash
./gcsls
```

## Editor setup

### Neovim (example)

```lua
vim.api.nvim_create_autocmd("FileType", {
  pattern = { "gcsim", "gcsl", "text" },
  callback = function()
    vim.lsp.start({
      name = "gcsls",
      cmd = { "/path/to/gcsls" },
      root_dir = vim.fn.getcwd(),
    })
  end,
})
```

Associate configs with a filetype, e.g. `*.gcsim` → `gcsim`.

### Zed

A local extension lives in `editors/zed/`. After installing `gcsls` on your `PATH`:

1. Command palette → **`zed: install dev extension`**
2. Select `editors/zed`
3. Optional: set `lsp.gcsls.binary.path` in `~/.config/zed/settings.json`

See [editors/zed/README.md](../../editors/zed/README.md).

### VS Code / other

Point any generic LSP client at the `gcsls` binary with stdio transport. There is no dedicated marketplace extension yet.

## Library

Reusable pieces live in [`pkg/gcs/lsp`](../../pkg/gcs/lsp):

- `Analyze(src)` — diagnostics from the parser
- `IndexDocument` / `DefinitionAt` / `DocumentSymbols` — local symbols
- `Complete` / `HoverAt` — completion and hover
- `RunStdio` — stdio host used by this command
