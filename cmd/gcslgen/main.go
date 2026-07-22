package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/genshinsim/gcsim/pkg/gcs/ast"
	"github.com/genshinsim/gcsim/pkg/gcs/eval"
	"github.com/genshinsim/gcsim/pkg/shortcut"
)

type langInfo struct {
	Name        string
	FileTypes   []string
	Keywords    []string
	CtrlFlow    []string
	ConfigKw    []string
	StatKeys    []string
	EleKeys     []string
	ActionNames []string
	CharNames   []string
	WeaponNames []string
	SetNames    []string
	BuiltinFns  []string
}

func run() error {
	info := &langInfo{
		Name:      "gcsim",
		FileTypes: []string{".gcsim", ".gcsl"},
	}

	// Extract from authoritative Go sources
	for _, kw := range ast.Keywords() {
		info.Keywords = append(info.Keywords, kw)
		switch kw {
		case "let", "if", "else", "while", "for", "switch", "case", "default",
			"break", "continue", "fallthrough", "return", "fn":
			info.CtrlFlow = append(info.CtrlFlow, kw)
		default:
			info.ConfigKw = append(info.ConfigKw, kw)
		}
	}

	for s := range ast.StatKeys {
		info.StatKeys = append(info.StatKeys, s)
	}
	for s := range ast.EleKeys {
		info.EleKeys = append(info.EleKeys, s)
	}
	for s := range ast.ActionKeys {
		info.ActionNames = append(info.ActionNames, s)
	}
	for s := range shortcut.CharNameToKey {
		info.CharNames = append(info.CharNames, s)
	}
	for s := range shortcut.WeaponNameToKey {
		info.WeaponNames = append(info.WeaponNames, s)
	}
	for s := range shortcut.SetNameToKey {
		info.SetNames = append(info.SetNames, s)
	}
	info.BuiltinFns = eval.BuiltinSysFuncs

	sort.Strings(info.Keywords)
	sort.Strings(info.CtrlFlow)
	sort.Strings(info.ConfigKw)
	sort.Strings(info.StatKeys)
	sort.Strings(info.EleKeys)
	sort.Strings(info.ActionNames)
	sort.Strings(info.CharNames)
	sort.Strings(info.WeaponNames)
	sort.Strings(info.SetNames)
	sort.Strings(info.BuiltinFns)

	// Output directory
	base := "editors/vscode"
	os.MkdirAll(base+"/syntaxes", 0755)
	os.MkdirAll(base+"/snippets", 0755)

	if err := writeTmGrammar(info, base); err != nil {
		return err
	}
	if err := writeVSCodeSnippets(info, base); err != nil {
		return err
	}
	if err := writeLangConfig(base); err != nil {
		return err
	}
	if err := writePackageJSON(info, base); err != nil {
		return err
	}

	fmt.Printf("generated %s/ from Go sources\n", base)
	fmt.Printf("  keywords: %d, actions: %d, stats: %d, elements: %d\n", len(info.Keywords), len(info.ActionNames), len(info.StatKeys), len(info.EleKeys))
	fmt.Printf("  characters: %d, weapons: %d, sets: %d, builtins: %d\n", len(info.CharNames), len(info.WeaponNames), len(info.SetNames), len(info.BuiltinFns))
	return nil
}

// --- TextMate grammar ---

type tmRoot struct {
	ScopeName  string                `json:"scopeName"`
	Name       string                `json:"name"`
	FileTypes  []string              `json:"fileTypes"`
	Patterns   []tmPattern           `json:"patterns"`
	Repository map[string]tmPattern  `json:"repository"`
}

type tmPattern struct {
	Name     string       `json:"name,omitempty"`
	Match    string       `json:"match,omitempty"`
	Begin    string       `json:"begin,omitempty"`
	End      string       `json:"end,omitempty"`
	Patterns []tmPattern  `json:"patterns,omitempty"`
	Include  string       `json:"include,omitempty"`
}

func writeTmGrammar(info *langInfo, base string) error {
	ctrlAlt := joinRE(info.CtrlFlow)
	configAlt := joinRE(info.ConfigKw)
	statAlt := joinRE(info.StatKeys)
	eleAlt := joinRE(info.EleKeys)
	actionAlt := joinRE(info.ActionNames)
	fnAlt := joinRE(info.BuiltinFns)

	g := tmRoot{
		ScopeName: "source.gcsim",
		Name:      "gcsim",
		FileTypes: []string{"gcsim", "gcsl"},
		Patterns: []tmPattern{
			{Include: "#comments"},
			{Include: "#strings"},
			{Include: "#numbers"},
			{Include: "#keywords"},
			{Include: "#operators"},
			{Include: "#builtins"},
			{Include: "#actions"},
			{Include: "#domain-words"},
			{Include: "#identifiers"},
		},
		Repository: map[string]tmPattern{
			"comments": {
				Patterns: []tmPattern{
					{Name: "comment.line.number-sign", Match: `#.*`},
					{Name: "comment.line.double-slash", Match: `//.*`},
				},
			},
			"strings": {
				Name:  "string.quoted.double",
				Begin: `"`,
				End:   `"`,
				Patterns: []tmPattern{
					{Name: "constant.character.escape", Match: `\\.`},
				},
			},
			"numbers": {
				Name:  "constant.numeric",
				Match: `\b(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)\b`,
			},
			"keywords": {
				Patterns: []tmPattern{
					{Name: "keyword.control", Match: `\b(?:` + ctrlAlt + `)\b`},
					{Name: "keyword.config",  Match: `\b(?:` + configAlt + `)\b`},
				},
			},
			"operators": {
				Name:  "keyword.operator",
				Match: `&&|\|\||==|!=|<=|>=|<|>|[+\-*/!]=?`,
			},
			"builtins": {
				Name:  "support.function",
				Match: `\b(?:` + fnAlt + `)\b`,
			},
			"actions": {
				Name:  "entity.name.function",
				Match: `\b(?:` + actionAlt + `)\b`,
			},
			"domain-words": {
				Patterns: []tmPattern{
					{Name: "variable.parameter.stat", Match: `\b(?:` + statAlt + `)\b`},
					{Name: "variable.parameter.element", Match: `\b(?:` + eleAlt + `)\b`},
				},
			},
			"identifiers": {
				Name:  "variable",
				Match: `\b[a-zA-Z_%][a-zA-Z0-9_%\-]*\b`,
			},
		},
	}
	return writeJSON(base+"/syntaxes/gcsl.tmLanguage.json", g)
}

// --- VS Code snippets ---

type snippet struct {
	Prefix      string   `json:"prefix"`
	Body        []string `json:"body"`
	Description string   `json:"description"`
}

func writeVSCodeSnippets(info *langInfo, base string) error {
	snips := map[string]snippet{
		"char-setup": {
			Prefix: "char", Body: []string{
				`${1:name} char lvl=${2:90/90} cons=${3:0} talent=${4:9,9,9};`,
				`${1:name} add weapon="${5:weapon}" refine=${6:5} lvl=90/90;`,
				`${1:name} add set="${7:set}" count=4;`,
				`${1:name} add stats hp=${8:4780} atk=${9:311} ${10};`,
			}, Description: "Character setup block",
		},
		"add-stats": {
			Prefix: "stats", Body: []string{
				`${1:name} add stats hp=${2} atk=${3} ${4};`,
			}, Description: "Add stats to character",
		},
		"for-loop": {
			Prefix: "for", Body: []string{
				`for let ${1:i}=${2:0}; ${1:i}<${3:n}; ${1:i}=${1:i}+1 {`,
				`\t${4}`,
				`}`,
			}, Description: "For loop",
		},
		"if": {
			Prefix: "if", Body: []string{
				`if ${1:condition} {`,
				`\t${2}`,
				`}`,
			}, Description: "If statement",
		},
		"ifelse": {
			Prefix: "ife", Body: []string{
				`if ${1:condition} {`,
				`\t${2}`,
				`} else {`,
				`\t${3}`,
				`}`,
			}, Description: "If/else statement",
		},
		"while": {
			Prefix: "while", Body: []string{
				`while ${1:condition} {`,
				`\t${2}`,
				`}`,
			}, Description: "While loop",
		},
		"fn": {
			Prefix: "fn", Body: []string{
				`fn ${1:name}(${2:args}) {`,
				`\t${3}`,
				`}`,
			}, Description: "Function declaration",
		},
		"let": {
			Prefix: "let", Body: []string{
				`let ${1:name} = ${2:value};`,
			}, Description: "Variable declaration",
		},
		"options": {
			Prefix: "options", Body: []string{
				`options iteration=${1:1000} duration=${2:90} ${3};`,
			}, Description: "Options block",
		},
		"target": {
			Prefix: "target", Body: []string{
				`target lvl=${1:100} resist=${2:0.1} hp=${3:999999999};`,
			}, Description: "Target configuration",
		},
		"energy": {
			Prefix: "energy", Body: []string{
				`energy every interval=${1:480},${2:720} amount=${3:1};`,
			}, Description: "Energy settings",
		},
		"action-chain": {
			Prefix: "chain", Body: []string{
				`${1:name} ${2:attack}, ${3:skill}, ${4:burst};`,
			}, Description: "Character action chain",
		},
		"switch": {
			Prefix: "switch", Body: []string{
				`switch ${1:expr} {`,
				`\tcase ${2:val}:`,
				`\t\t${3}`,
				`\tdefault:`,
				`\t\t${4}`,
				`}`,
			}, Description: "Switch statement",
		},
	}
	return writeJSON(base+"/snippets/gcsl.json", snips)
}

// --- VS Code language config ---

type langConfig struct {
	Comments          commentConfig     `json:"comments"`
	Brackets          [][]string        `json:"brackets"`
	AutoClosingPairs  [][]string        `json:"autoClosingPairs"`
	SurroundingPairs  [][]string        `json:"surroundingPairs"`
	IndentationRules  indentRules       `json:"indentationRules"`
}

type commentConfig struct {
	LineComment string   `json:"lineComment"`
}

type indentRules struct {
	IncreaseIndent string `json:"increaseIndentPattern"`
	DecreaseIndent string `json:"decreaseIndentPattern"`
}

func writeLangConfig(base string) error {
	cfg := langConfig{
		Comments: commentConfig{LineComment: "#"},
		Brackets: [][]string{
			{"{", "}"}, {"[", "]"}, {"(", ")"},
		},
		AutoClosingPairs: [][]string{
			{"{", "}"}, {"[", "]"}, {"(", ")"}, {"\"", "\""},
		},
		SurroundingPairs: [][]string{
			{"{", "}"}, {"[", "]"}, {"(", ")"}, {"\"", "\""},
		},
		IndentationRules: indentRules{
			IncreaseIndent: `(\{[^}"']*|\([^)"']*|\[[^\]"']*)\s*$`,
			DecreaseIndent: `^\s*(\}|\]|\))`,
		},
	}
	return writeJSON(base+"/language-configuration.json", cfg)
}

// --- VS Code extension manifest ---

type vpkg struct {
	Name            string              `json:"name"`
	DisplayName     string              `json:"displayName"`
	Description     string              `json:"description"`
	Version         string              `json:"version"`
	Publisher       string              `json:"publisher"`
	Engines         map[string]string   `json:"engines"`
	Categories      []string            `json:"categories"`
	Main            string              `json:"main,omitempty"`
	ActivationEvts  []string            `json:"activationEvents,omitempty"`
	Deps            map[string]string   `json:"dependencies,omitempty"`
	License         string              `json:"license,omitempty"`
	Repo            *vrepo              `json:"repository,omitempty"`
	Bugs            *vbugs              `json:"bugs,omitempty"`
	Homepage        string              `json:"homepage,omitempty"`
	Scripts         map[string]string   `json:"scripts,omitempty"`
	Contributes     vcontributes        `json:"contributes"`
}

type vrepo struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type vbugs struct {
	URL string `json:"url"`
}

type vcontributes struct {
	Languages     []vlang          `json:"languages"`
	Grammars      []vgrammar       `json:"grammars"`
	Snippets      []vsnipref       `json:"snippets"`
	Configuration *vconfigSection  `json:"configuration,omitempty"`
}

type vconfigSection struct {
	Title      string              `json:"title"`
	Properties map[string]vprop    `json:"properties"`
}

type vprop struct {
	Type        string `json:"type"`
	Default     any    `json:"default"`
	Description string `json:"description"`
}

type vlang struct {
	ID            string   `json:"id"`
	Aliases       []string `json:"aliases"`
	Extensions    []string `json:"extensions"`
	Configuration string   `json:"configuration"`
}

type vgrammar struct {
	Language  string `json:"language"`
	ScopeName string `json:"scopeName"`
	Path      string `json:"path"`
}

type vsnipref struct {
	Language string `json:"language"`
	Path     string `json:"path"`
}

func writePackageJSON(info *langInfo, base string) error {
	pkg := vpkg{
		Name:        "gcsl",
		DisplayName: "gcsim Language Support",
		Description: "Syntax highlighting, snippets, and language server support for gcsim config files (gcsl)",
		Version:     "0.1.0",
		Publisher:   "gcsim",
		Engines:     map[string]string{"vscode": "^1.84.0"},
		Categories:  []string{"Programming Languages", "Snippets", "Linters"},
			Main:        "./extension.js",
		ActivationEvts: []string{"onLanguage:gcsim"},
		Deps:        map[string]string{"vscode-languageclient": "^9.0.1"},
		License:     "MIT",
		Repo: &vrepo{
			Type: "git",
			URL:  "https://github.com/genshinsim/gcsim.git",
		},
		Bugs:     &vbugs{URL: "https://github.com/genshinsim/gcsim/issues"},
		Homepage: "https://gcsim.app",
		Scripts: map[string]string{
			"vscode:prepublish": "go run ../../cmd/gcslgen",
		},
		Contributes: vcontributes{
			Languages: []vlang{{
				ID: "gcsim", Aliases: []string{"gcsim", "gcsl"},
				Extensions: []string{".gcsim", ".gcsl"},
				Configuration: "./language-configuration.json",
			}},
			Grammars: []vgrammar{{
				Language: "gcsim", ScopeName: "source.gcsim",
				Path: "./syntaxes/gcsl.tmLanguage.json",
			}},
			Snippets: []vsnipref{{
				Language: "gcsim", Path: "./snippets/gcsl.json",
			}},
			Configuration: &vconfigSection{
				Title: "gcsls",
				Properties: map[string]vprop{
					"gcsls.path": {
						Type: "string", Default: "gcsls",
						Description: "Path to the gcsls language server binary. Default: look up on PATH.",
					},
				},
			},
		},
	}
	return writeJSON(base+"/package.json", pkg)
}

// --- helpers ---

func joinRE(items []string) string {
	esc := make([]string, len(items))
	for i, s := range items {
		esc[i] = regexp.QuoteMeta(s)
	}
	return strings.Join(esc, "|")
}

func writeJSON(path string, v any) error {
	dir := filepath.Dir(path)
	os.MkdirAll(dir, 0755)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
