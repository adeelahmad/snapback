package resolver

// DirectoryKey returns the stable 32-hex-character key for rel under rootID.
func DirectoryKey(rootID, rel string) string {
	panic("SUB-AGENT-TODO: T2 hex(sha256(AppendUvarint(len(rootID))||rootID||AppendUvarint(len(rel))||rel))[:32]")
}
