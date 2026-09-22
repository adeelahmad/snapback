package web

import "net/url"

// storeCredential writes a typed repository password to a credential file.
func storeCredential(stateDir, repoID, secret string) (string, error) {
	return "", nil
}

// credentialMode reports how repository i of the form supplies its password.
func credentialMode(v url.Values, i int) string {
	return ""
}
