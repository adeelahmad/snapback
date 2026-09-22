package shellhook

import "embed"

//go:embed snippets
var snippets embed.FS

// Script returns the hook snippet for shell, one of bash, zsh or fish.
func Script(shell string) (string, error) {
	panic("SUB-AGENT-TODO: read snippets/snapback.<shell> from the embedded snippets FS for bash; unknown shell returns an error naming bash|zsh|fish; bash snippet chains PROMPT_COMMAND (string or array), preserves $?, dedups on physical dir, detaches notify")
}
