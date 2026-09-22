package shellhook

import (
	"embed"
	"fmt"
)

//go:embed snippets
var snippets embed.FS

// Script returns the hook snippet for shell, one of bash, zsh or fish.
func Script(shell string) (string, error) {
	switch shell {
	case "bash", "zsh", "fish":
	default:
		return "", fmt.Errorf("unsupported shell %q: want bash|zsh|fish", shell)
	}
	b, err := snippets.ReadFile("snippets/snapback." + shell)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
