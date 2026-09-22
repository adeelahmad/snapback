package resolver

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
)

// DirectoryKey returns the stable 32-hex-character key for rel under rootID.
func DirectoryKey(rootID, rel string) string {
	var buf []byte
	buf = binary.AppendUvarint(buf, uint64(len(rootID)))
	buf = append(buf, rootID...)
	buf = binary.AppendUvarint(buf, uint64(len(rel)))
	buf = append(buf, rel...)
	sum := sha256.Sum256(buf)
	return hex.EncodeToString(sum[:])[:32]
}
