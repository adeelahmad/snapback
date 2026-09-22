package web

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/adeelahmad/snapback/internal/config"
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

// FormSections renders cfg as the basic/advanced split the form shows: the
// controls of each group, with every error anchored to its own control, and
// the advanced badge counting the fields Split hid, not the controls the
// password mode adds.
func FormSections(cfg *config.Config, byPath map[string]string, form url.Values) webui.FormSections {
	basic, advanced := Split(Fields(cfg))
	return webui.FormSections{
		Basic:         groupSections(basic.Fields, byPath, form),
		Advanced:      groupSections(advanced.Fields, byPath, form),
		AdvancedCount: advanced.Count,
	}
}

// groupSections turns one group's fields into its sections, keeping the
// password controls attached to the repository's password file.
func groupSections(fields []webui.Field, byPath map[string]string, form url.Values) []webui.ConfigSection {
	var out []webui.ConfigSection
	for _, sec := range webui.Sections(fields) {
		s := webui.ConfigSection{Title: sec.Title}
		for _, f := range sec.Fields {
			s.Controls = append(s.Controls, control(f, byPath[f.Path]))
			if prefix, ok := strings.CutSuffix(f.Path, ".password_file"); ok {
				s.Controls = append(s.Controls, passwordControls(prefix, byPath, form)...)
			}
		}
		out = append(out, s)
	}
	return out
}
