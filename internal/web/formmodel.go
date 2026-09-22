package web

import (
	"errors"
	"fmt"
	"maps"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/webui"
)

var durationType = reflect.TypeFor[time.Duration]()

// formSelects lists the choices of every key rendered as a select, keyed by
// its key path with the slice indices removed.
var formSelects = map[string][]webui.Option{
	"timestamps": {
		{Value: "utc", Label: "UTC"},
		{Value: "local", Label: "Local"},
	},
	"repositories.lock_mode": {
		{Value: "normal", Label: "Normal"},
		{Value: "none", Label: "None"},
	},
	"discovery.mode": {
		{Value: "seed", Label: "Seed"},
		{Value: "on-access", Label: "On access"},
	},
	"service.manager": {
		{Value: "auto", Label: "Auto"},
		{Value: "systemd", Label: "systemd"},
	},
	"service.scope": {
		{Value: "user", Label: "User"},
		{Value: "system", Label: "System"},
	},
}

// maxFormIndex bounds the slice index a form value may address, so a crafted
// name cannot allocate an unbounded slice.
const maxFormIndex = 1000

// Decode builds a configuration from form values whose names are YAML key
// paths, such as repositories[0].lock_mode. Every value it cannot apply is
// reported as a field error instead of being dropped.
func Decode(v url.Values) (*config.Config, []config.FieldError) {
	cfg := &config.Config{}
	root := reflect.ValueOf(cfg).Elem()
	var errs []config.FieldError
	for _, name := range slices.Sorted(maps.Keys(v)) {
		if err := assignPath(root, name, v[name]); err != nil {
			errs = append(errs, config.FieldError{Path: name, Msg: err.Error()})
		}
	}
	return cfg, errs
}

// Fields lists every configuration key of cfg as a renderable form field, in
// declaration order, with a row marker before each repeatable element.
func Fields(cfg *config.Config) []webui.Field {
	var out []webui.Field
	appendFields(&out, "", reflect.ValueOf(cfg).Elem())
	return out
}

func appendFields(out *[]webui.Field, prefix string, v reflect.Value) {
	t := v.Type()
	for i := range t.NumField() {
		tag := t.Field(i).Tag.Get("yaml")
		if tag == "" {
			continue
		}
		path := tag
		if prefix != "" {
			path = prefix + "." + tag
		}
		fv := v.Field(i)
		switch {
		case fv.Kind() == reflect.Slice && fv.Type().Elem().Kind() == reflect.Struct:
			appendRows(out, path, tag, fv)
		case fv.Kind() == reflect.Struct:
			appendFields(out, path, fv)
		default:
			*out = append(*out, leafField(path, fv))
		}
	}
}

// appendRows renders one row per element, and an empty prototype row when
// there is none, so every key stays reachable in a blank form.
func appendRows(out *[]webui.Field, path, tag string, v reflect.Value) {
	n := max(v.Len(), 1)
	for i := range n {
		row := fmt.Sprintf("%s[%d]", path, i)
		*out = append(*out, webui.Field{
			Path:  row,
			Kind:  webui.KindRow,
			Label: fmt.Sprintf("%s %d", fieldLabel(tag), i+1),
		})
		elem := reflect.Zero(v.Type().Elem())
		if i < v.Len() {
			elem = v.Index(i)
		}
		appendFields(out, row, elem)
	}
}

func leafField(path string, v reflect.Value) webui.Field {
	f := webui.Field{Path: path, Label: fieldLabel(lastSegment(path))}
	if opts, ok := formSelects[unindexed(path)]; ok {
		f.Kind, f.Options, f.Value = webui.KindSelect, opts, v.String()
		return f
	}
	switch {
	case v.Type() == durationType:
		f.Kind, f.Value = webui.KindDuration, time.Duration(v.Int()).String()
	case v.Kind() == reflect.String:
		f.Kind, f.Value = webui.KindText, v.String()
	case v.Kind() == reflect.Bool:
		f.Kind, f.Value = webui.KindToggle, strconv.FormatBool(v.Bool())
	case v.CanInt():
		f.Kind, f.Value = webui.KindNumber, strconv.FormatInt(v.Int(), 10)
	case v.CanFloat():
		f.Kind, f.Value = webui.KindNumber, strconv.FormatFloat(v.Float(), 'g', -1, 64)
	case v.Kind() == reflect.Slice:
		f.Kind = webui.KindChips
		for i := range v.Len() {
			f.Values = append(f.Values, v.Index(i).String())
		}
	case v.Kind() == reflect.Map:
		f.Kind = webui.KindChips
		for _, k := range slices.Sorted(maps.Keys(v.Interface().(map[string]string))) {
			f.Values = append(f.Values, k+"="+v.MapIndex(reflect.ValueOf(k)).String())
		}
	}
	return f
}

// assignPath walks path through root and applies vals to the value it names.
func assignPath(root reflect.Value, path string, vals []string) error {
	cur := root
	for _, seg := range strings.Split(path, ".") {
		name, indices, err := splitSegment(seg)
		if err != nil {
			return err
		}
		if cur.Kind() != reflect.Struct {
			return errors.New("unknown field")
		}
		cur, err = fieldByYAML(cur, name)
		if err != nil {
			return err
		}
		for _, i := range indices {
			if cur.Kind() != reflect.Slice || cur.Type().Elem().Kind() != reflect.Struct {
				return errors.New("unknown field")
			}
			if i >= maxFormIndex {
				return fmt.Errorf("index must be below %d", maxFormIndex)
			}
			for cur.Len() <= i {
				cur.Set(reflect.Append(cur, reflect.Zero(cur.Type().Elem())))
			}
			cur = cur.Index(i)
		}
	}
	return setLeaf(cur, vals)
}

// splitSegment splits name[1][2] into its name and its indices.
func splitSegment(seg string) (string, []int, error) {
	name, rest, found := strings.Cut(seg, "[")
	if name == "" {
		return "", nil, errors.New("unknown field")
	}
	if !found {
		return name, nil, nil
	}
	var indices []int
	for _, part := range strings.Split(rest, "[") {
		digits, ok := strings.CutSuffix(part, "]")
		if !ok {
			return "", nil, errors.New("malformed field index")
		}
		i, err := strconv.Atoi(digits)
		if err != nil || i < 0 {
			return "", nil, errors.New("malformed field index")
		}
		indices = append(indices, i)
	}
	return name, indices, nil
}

func fieldByYAML(v reflect.Value, name string) (reflect.Value, error) {
	t := v.Type()
	for i := range t.NumField() {
		if t.Field(i).Tag.Get("yaml") == name {
			return v.Field(i), nil
		}
	}
	return reflect.Value{}, errors.New("unknown field")
}

func setLeaf(v reflect.Value, vals []string) error {
	switch {
	case v.Kind() == reflect.Slice && v.Type().Elem().Kind() == reflect.String:
		if len(vals) == 0 {
			return nil
		}
		v.Set(reflect.ValueOf(slices.Clone(vals)))
		return nil
	case v.Kind() == reflect.Map:
		return setMap(v, vals)
	case v.Kind() == reflect.Struct, v.Kind() == reflect.Slice:
		return errors.New("unknown field")
	}
	if len(vals) == 0 {
		return nil
	}
	return setScalar(v, vals[0])
}

func setScalar(v reflect.Value, s string) error {
	switch {
	case v.Type() == durationType:
		d, err := time.ParseDuration(s)
		if err != nil {
			return errors.New("must be a duration such as 60s")
		}
		v.SetInt(int64(d))
	case v.Kind() == reflect.String:
		v.SetString(s)
	case v.Kind() == reflect.Bool:
		switch strings.ToLower(s) {
		case "on", "true", "1", "yes":
			v.SetBool(true)
		case "", "off", "false", "0", "no":
			v.SetBool(false)
		default:
			return errors.New("must be on or off")
		}
	case v.CanInt():
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return errors.New("must be a whole number")
		}
		v.SetInt(n)
	case v.CanFloat():
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return errors.New("must be a number")
		}
		v.SetFloat(f)
	default:
		return errors.New("unknown field")
	}
	return nil
}

func setMap(v reflect.Value, vals []string) error {
	if v.Type().Key().Kind() != reflect.String || v.Type().Elem().Kind() != reflect.String {
		return errors.New("unknown field")
	}
	m := map[string]string{}
	for _, s := range vals {
		k, val, ok := strings.Cut(s, "=")
		if !ok || k == "" {
			return errors.New("must be KEY=VALUE")
		}
		m[k] = val
	}
	if len(m) == 0 {
		return nil
	}
	v.Set(reflect.ValueOf(m))
	return nil
}

// unindexed strips the slice indices from a key path, so that
// repositories[2].lock_mode keys the same entry as repositories[0].lock_mode.
func unindexed(path string) string {
	var b strings.Builder
	for _, seg := range strings.Split(path, ".") {
		name, _, _ := strings.Cut(seg, "[")
		if b.Len() > 0 {
			b.WriteByte('.')
		}
		b.WriteString(name)
	}
	return b.String()
}

func lastSegment(path string) string {
	if i := strings.LastIndexByte(path, '.'); i >= 0 {
		return path[i+1:]
	}
	return path
}

// fieldLabel turns a yaml key such as lock_mode into the label "Lock mode".
func fieldLabel(tag string) string {
	s := strings.ReplaceAll(tag, "_", " ")
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
