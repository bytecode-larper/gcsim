# gcsl — gcsim Language Support

Syntax highlighting, snippets, and LSP-powered IDE features for [gcsim](https://gcsim.app) config files (`.gcsim`, `.gcsl`).

The language server (`gcsls`) is bundled with the extension — no separate install needed.

## Features

- **Syntax highlighting** — keywords, operators, strings, numbers, stats, elements, actions
- **Snippets** — character setup, for loops, if/else, options, target, energy, action chains
- **Diagnostics** — real-time syntax and config validation via the gcsl PEG parser
- **Completions** — characters, actions, weapons, artifact sets, stats, elements, system functions, user-defined functions
- **Hover docs** — documentation for all of the above
- **Go to definition** — on user functions and variables

## Requirements

No additional installation required. The language server is included for Linux x64, Linux arm64, macOS x64, macOS arm64, and Windows x64.

To use a custom build of `gcsls` (e.g. a newer version), set the path in VS Code settings:

```json
{
  "gcsls.path": "/path/to/your/gcsls"
}
```

## Credits

Built on [pigeon](https://github.com/mna/pigeon) PEG parser and [go.lsp.dev](https://github.com/go-language-server/protocol).
