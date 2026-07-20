use zed_extension_api::{
    self as zed, serde_json, settings::LspSettings, LanguageServerId, Result, Worktree,
};

struct GcsimExtension;

impl zed::Extension for GcsimExtension {
    fn new() -> Self {
        Self
    }

    fn language_server_command(
        &mut self,
        language_server_id: &LanguageServerId,
        worktree: &Worktree,
    ) -> Result<zed::Command> {
        match language_server_id.as_ref() {
            "gcsls" => resolve_gcsls(worktree),
            other => Err(format!("unknown language server: {other}")),
        }
    }

    fn language_server_initialization_options(
        &mut self,
        language_server_id: &LanguageServerId,
        worktree: &Worktree,
    ) -> Result<Option<serde_json::Value>> {
        Ok(LspSettings::for_worktree(language_server_id.as_ref(), worktree)
            .ok()
            .and_then(|s| s.initialization_options))
    }

    fn language_server_workspace_configuration(
        &mut self,
        language_server_id: &LanguageServerId,
        worktree: &Worktree,
    ) -> Result<Option<serde_json::Value>> {
        Ok(LspSettings::for_worktree(language_server_id.as_ref(), worktree)
            .ok()
            .and_then(|s| s.settings))
    }
}

fn resolve_gcsls(worktree: &Worktree) -> Result<zed::Command> {
    let mut args: Vec<String> = Vec::new();
    let env = worktree.shell_env();

    // Prefer an explicit binary from settings:
    //   "lsp": { "gcsls": { "binary": { "path": "...", "arguments": [] } } }
    if let Ok(lsp_settings) = LspSettings::for_worktree("gcsls", worktree) {
        if let Some(binary) = lsp_settings.binary {
            if let Some(arguments) = binary.arguments {
                args = arguments;
            }
            if let Some(path) = binary.path {
                return Ok(zed::Command {
                    command: path,
                    args,
                    env,
                });
            }
        }
    }

    // Fall back to PATH (worktree shell env / which).
    if let Some(path) = worktree.which("gcsls") {
        return Ok(zed::Command {
            command: path,
            args,
            env,
        });
    }

    Err(
        "gcsls not found. Install with `go install`/`go build -o ~/.local/bin/gcsls ./cmd/gcsls`, \
         or set lsp.gcsls.binary.path in Zed settings."
            .into(),
    )
}

zed::register_extension!(GcsimExtension);
