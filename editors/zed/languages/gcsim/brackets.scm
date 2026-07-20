("{" @open "}" @close)
("[" @open "]" @close)
("(" @open ")" @close)
; Quotes are not separate tree nodes: string is a single token() leaf in the grammar.
; Auto-close for " is handled via config.toml brackets instead.