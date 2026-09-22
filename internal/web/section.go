package web

import (
	"regexp"

	"github.com/adeelahmad/snapback/internal/webui"
)

// Group is one half of the configuration form: the fields it shows and how
// many they are, so a template can label a collapsed group without counting.
type Group struct {
	Fields []webui.Field
	Count  int
}

// index matches a list index so paths of any element compare as one key.
var index = regexp.MustCompile(`\[\d+\]`)

// basicKeys are the indexless paths of the few keys a first run has to answer:
// where the repository is, how it is unlocked, and which directory is saved.
var basicKeys = map[string]bool{
	"repositories[].repository":    true,
	"repositories[].password_file": true,
	"roots[].local_path":           true,
}

// Split partitions the configuration fields into the few a first run has to
// answer and everything else, keeping the order of fields.
func Split(fields []webui.Field) (basic, advanced Group) {
	for _, f := range fields {
		if basicKeys[index.ReplaceAllString(f.Path, "[]")] {
			basic.Fields = append(basic.Fields, f)
			continue
		}
		advanced.Fields = append(advanced.Fields, f)
	}
	basic.Count, advanced.Count = len(basic.Fields), len(advanced.Fields)
	return basic, advanced
}
