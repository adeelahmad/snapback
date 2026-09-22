// agentic:shim
package shellhook

// Script is a compile shim with a deliberately wrong body.
func Script(shell string) (string, error) {
	return "echo $x\n", nil
}
