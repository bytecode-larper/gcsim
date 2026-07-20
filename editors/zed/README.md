# Zed extension for gcsim / gcsl

Registers the **gcsim** language (`.gcsim`, `.gcsl`) and starts the **gcsls** language server.

## Prerequisites

1. Build and install `gcsls` somewhere on your `PATH` (e.g. `~/.local/bin`):

```bash
# from the gcsim repo root
go build -o ~/.local/bin/gcsls ./cmd/gcsls
```

Or pin an absolute path in Zed settings (see below).

## Install this extension in Zed

1. Open the command palette (`Ctrl+Shift+P`).
2. Run **`zed: install dev extension`**.
3. Select this directory: `editors/zed` (inside the gcsim repo).

Zed will build the Wasm extension, compile the Tree-sitter grammar, and enable it.

### After grammar / highlight changes

Highlight queries live in **one** place: `tree-sitter-gcsim/queries/highlights.scm`.  
`languages/gcsim/highlights.scm` is a symlink to that file (do not duplicate).

1. Commit updates inside `tree-sitter-gcsim/` (it is a small local git repo so Zed can fetch it):

   ```bash
   cd editors/zed/tree-sitter-gcsim
   tree-sitter generate
   git add -A && git commit -m "update grammar"
   # put the new SHA into ../extension.toml [grammars.gcsim].rev
   ```

2. Re-run **`zed: install dev extension`** on `editors/zed` (or uninstall + reinstall the dev extension).
3. Reload Zed / reopen the `.gcsim` buffer.

You should see colors for comments (`#`, `//`), keywords (`fn`, `while`, `for`, `if`, …), function names, actions, strings, and numbers.

## Optional settings (`~/.config/zed/settings.json`)

```json
{
  "lsp": {
    "gcsls": {
      "binary": {
        "path": "/home/YOU/.local/bin/gcsls",
        "arguments": []
      }
    }
  },
  "languages": {
    "gcsim": {
      "language_servers": ["gcsls"]
    }
  },
  "file_types": {
    "gcsim": ["*.gcsim", "*.gcsl"]
  }
}
```

`file_types` is usually unnecessary if `path_suffixes` in `languages/gcsim/config.toml` already match.

## Verify

1. Open a file ending in `.gcsim` (or set the buffer language to **gcsim**).
2. Open **`zed: open language server logs`** from the palette if diagnostics do not appear.
3. Type an invalid line such as `let x = ;` — you should get a red diagnostic from `gcsls`.
4. On `editors/zed/examples/sample.gcsim`: hover a character/weapon, go-to-definition on `mualani_combo2()`, and check the outline for user functions.

After rebuilding `gcsls`, restart the language server (or reload Zed) so the editor picks up the new binary.
