package lsp

// Documentation overlays for identifiers. Names for keywords/actions/sysfuncs/stats/elements
// come from the real language tables (ast, eval, shortcut). These maps only supply prose;
// missing entries still complete/hover with a generic label.

var keywordDocs = map[string]string{
	"let":                 "Declare a local number variable: `let name = expr;`",
	"while":               "Loop while condition is true: `while cond { ... }`",
	"if":                  "Conditional: `if cond { ... } else { ... }`",
	"else":                "Else branch of an `if`",
	"fn":                  "Declare a function: `fn name(args) { ... }`",
	"switch":              "Switch on a value: `switch expr { case x: ... default: ... }`",
	"case":                "Case arm inside `switch`",
	"default":             "Default arm inside `switch`",
	"break":               "Exit the innermost loop or switch",
	"continue":            "Continue to the next loop iteration",
	"fallthrough":         "Fall through to the next switch case",
	"return":              "Return from a function: `return` or `return expr;`",
	"for":                 "C-style loop: `for init; cond; post { ... }`",
	"options":             "Simulator options line: `options swap_delay=12 iteration=1000;`",
	"add":                 "Attach gear/stats to a character: `char add weapon=...` / `set=...` / `stats ...`",
	"char":                "Character setup: `name char lvl=... cons=... talent=...;`",
	"stats":               "Stat line: `char add stats hp=... atk=...;`",
	"weapon":              "Weapon assignment key: `weapon=\"name\"`",
	"set":                 "Artifact set key: `set=\"name\"`",
	"lvl":                 "Level key (`90/90` or number) for char/weapon",
	"refine":              "Weapon refinement (1–5)",
	"cons":                "Character constellation (0–6)",
	"talent":              "Talent levels: `talent=a,e,q`",
	"count":               "Number of set pieces: `count=4`",
	"params":              "Extra parameters map: `+params=[key=val]`",
	"label":               "Label for control flow (legacy)",
	"until":               "Until condition (legacy)",
	"active":              "Starting on-field character: `active name;`",
	"target":              "Enemy target line: `target lvl=... resist=...;`",
	"particle_threshold":  "Target particle drop threshold",
	"particle_drop_count": "Target particle drop count",
	"particle_element":    "Target particle element",
	"resist":              "Base resistance applied to all elements on a target",
	"energy":              "Energy regen schedule: `energy every interval=a,b amount=n;`",
	"hurt":                "Hurt schedule: `hurt every interval=a,b amount=...;`",
}

var actionDocs = map[string]string{
	"skill":       "Character elemental skill",
	"burst":       "Character elemental burst",
	"attack":      "Normal attack",
	"charge":      "Charged attack",
	"high_plunge": "High plunge attack",
	"low_plunge":  "Low plunge attack",
	"aim":         "Aimed shot (bow)",
	"dash":        "Dash",
	"jump":        "Jump",
	"walk":        "Walk (often with `[f=n]` frames)",
	"swap":        "Swap character",
}

var sysFuncDocs = map[string]string{
	"f":                        "Current simulation frame (system). Call as `f()`.",
	"rand":                     "Random number in [0,1). Call as `rand()`.",
	"randnorm":                 "Normally distributed random number",
	"print":                    "Print values to the sim log: `print(...)`",
	"wait":                     "Wait N frames: `wait(n)`",
	"sleep":                    "Alias of wait",
	"delay":                    "Delay helper",
	"type":                     "Type helper",
	"execute_action":           "Execute a raw action by key",
	"set_target_pos":           "Set target position",
	"set_player_pos":           "Set player position",
	"set_default_target":       "Set default target index",
	"set_swap_icd":             "Set swap ICD",
	"set_particle_delay":       "Set particle delay",
	"kill_target":              "Kill a target by index",
	"is_target_dead":           "Whether a target is dead",
	"pick_up_crystallize":      "Pick up crystallize shards",
	"set_starting_verdant_dew": "Set starting verdant dew",
	"sin":                      "Sine (radians)",
	"cos":                      "Cosine (radians)",
	"asin":                     "Arcsine",
	"acos":                     "Arccosine",
	"is_even":                  "Whether a number is even",
	"set_on_tick":              "Register an on-tick callback",
}

// optionKeys is the set of `options ...` assignment keys (from parseOptions).
// No shared runtime table exists yet; names live here with optional docs.
var optionDocs = map[string]string{
	"iteration":           "Number of Monte Carlo iterations",
	"duration":            "Max simulation duration in frames (if no other exit)",
	"swap_delay":          "Frames inserted on character swap",
	"workers":             "Parallel worker count",
	"hitlag":              "Enable hitlag simulation",
	"defhalt":             "Defhalt setting",
	"ignore_burst_energy": "Ignore burst energy requirements",
	"debug":               "Legacy debug flag (ignored; runs always log)",
	"mode":                "Legacy mode option",
	"attack_delay":        "Extra frames after attack",
	"charge_delay":        "Extra frames after charge",
	"skill_delay":         "Extra frames after skill",
	"burst_delay":         "Extra frames after burst",
	"jump_delay":          "Extra frames after jump",
	"dash_delay":          "Extra frames after dash",
	"aim_delay":           "Extra frames after aim",
	"frame_defaults":      "Preset delay profile (e.g. `human`)",
}

// fieldDocs covers common `.path` segments (conditional fields). Names are the catalog.
var fieldDocs = map[string]string{
	"energy":             "Current energy",
	"energymax":          "Max energy",
	"hp":                 "Current HP",
	"hpratio":            "Current HP ratio",
	"hpmax":              "Max HP",
	"bol":                "Bond of Life amount",
	"bolratio":           "Bond of Life ratio",
	"cons":               "Constellation level",
	"normal":             "Normal attack counter / state",
	"onfield":            "Whether character is on-field",
	"weapon":             "Weapon info field",
	"mods":               "Active mods",
	"status":             "Status map / duration",
	"infusion":           "Weapon infusion",
	"tags":               "Character tags",
	"sets":               "Equipped sets",
	"stats":              "Stat values",
	"skill":              "Skill subfields (e.g. `.char.skill.cd`)",
	"burst":              "Burst subfields (e.g. `.char.burst.ready`)",
	"attack":             "Attack subfields",
	"charge":             "Charge attack subfields",
	"cd":                 "Cooldown remaining",
	"ready":              "Ability ready (bool as number)",
	"nightsoul":          "Nightsoul blessing fields",
	"state":              "Nightsoul / mode state",
	"points":             "Nightsoul points",
	"duration":           "Remaining duration",
	"element":            "Element fields",
	"debuff":             "Debuff fields",
	"res":                "Resistance fields",
	"def":                "Defense fields",
	"stam":               "Stamina",
	"construct":          "Construct gadgets",
	"gadgets":            "Gadgets",
	"dendrocore":         "Dendro core gadgets",
	"sourcewaterdroplet": "Sourcewater droplet gadgets",
	"crystallizeshard":   "Crystallize shards",
	"keys":               "Key listing helpers",
	"action":             "Last / related action fields",
	"previous-char":      "Previous character",
	"previous-action":    "Previous action",
	"airborne":           "Airborne state",
	"momentum":           "Momentum stacks (e.g. Mualani)",
	"nightsoul.state":     "Whether nightsoul state is active",
}

// configKeyDocs: energy/hurt/target line keys that are not always lexer keywords.
var configKeyDocs = map[string]string{
	"every":         "Repeat schedule (energy/hurt): `every interval=a,b amount=...`",
	"once":          "One-shot schedule (energy/hurt): `once interval=n amount=...`",
	"interval":      "Frame interval: single value or `min,max` range",
	"amount":        "Amount for energy/hurt events",
	"pos":           "Target position: `pos=x,y`",
	"radius":        "Target hitbox radius",
	"freeze_resist": "Target freeze resistance",
	"type":          "Enemy type key (damage mode)",
	"hp":            "Target HP (enables damage mode when set on target)",
}

// actionParamDocs: common keys inside action `[...]` or `+params=[...]`.
var actionParamDocs = map[string]string{
	"f":          "Frame duration / cancel frames for an action",
	"hold":       "Hold length for skills/charges (character-specific)",
	"travel":     "Projectile travel frames",
	"hits":       "Number of hits",
	"n":          "Hit / sequence index (character-specific)",
	"stacks":     "Starting stacks (weapon/character param)",
	"hold_ticks": "Hold tick count",
	"src":        "Source index param",
	"delay":      "Extra delay frames",
}

func docOr(m map[string]string, key, fallback string) string {
	if d, ok := m[key]; ok && d != "" {
		return d
	}
	return fallback
}
