const vscode = require('vscode');
const { LanguageClient, TransportKind } = require('vscode-languageclient/node');
const path = require('path');
const fs = require('fs');

function serverBinary() {
  // 1. User-configured path wins
  const configured = vscode.workspace.getConfiguration('gcsls').get('path');
  if (configured && configured !== 'gcsls' && fs.existsSync(configured)) {
    return configured;
  }

  // 2. Bundled binary (platform-specific VSIX puts it here)
  const ext = process.platform === 'win32' ? '.exe' : '';
  const bundled = path.join(__dirname, 'server', `gcsls${ext}`);
  if (fs.existsSync(bundled)) {
    return bundled;
  }

  // 3. Fallback to PATH
  return 'gcsls';
}

function activate(context) {
  const command = serverBinary();

  const serverOptions = { command, transport: TransportKind.stdio };

  const clientOptions = {
    documentSelector: [{ scheme: 'file', language: 'gcsim' }],
    synchronize: { configurationSection: 'gcsls' },
  };

  const client = new LanguageClient('gcsls', 'gcsls', serverOptions, clientOptions);
  client.start();
  context.subscriptions.push({ dispose: () => client.stop() });
}

function deactivate() {}

module.exports = { activate, deactivate };
