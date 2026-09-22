package config

import (
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"
)

var snakeCaseTag = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// yamlName returns the name part of a field's yaml tag, without options.
func yamlName(f reflect.StructField) string {
	name, _, _ := strings.Cut(f.Tag.Get("yaml"), ",")
	return name
}

// checkYAMLTags reports every exported field of struct type typ, and of the
// struct types reachable from it through fields, slices, maps and pointers,
// whose yaml tag is not snake_case.
func checkYAMLTags(t *testing.T, typ reflect.Type, seen map[reflect.Type]bool) {
	t.Helper()
	for typ.Kind() == reflect.Pointer || typ.Kind() == reflect.Slice || typ.Kind() == reflect.Map {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct || seen[typ] || typ.PkgPath() != reflect.TypeFor[Config]().PkgPath() {
		return
	}
	seen[typ] = true
	for i := range typ.NumField() {
		f := typ.Field(i)
		if !f.IsExported() {
			continue
		}
		if name := yamlName(f); !snakeCaseTag.MatchString(name) {
			t.Errorf("%s.%s yaml tag name = %q, want snake_case matching %s", typ.Name(), f.Name, name, snakeCaseTag)
		}
		checkYAMLTags(t, f.Type, seen)
	}
}

func TestConfigYAMLTagsAreSnakeCase(t *testing.T) {
	typ := reflect.TypeFor[Config]()
	checkYAMLTags(t, typ, map[reflect.Type]bool{})

	want := []string{
		"version", "link_name", "timestamps", "state_dir", "history_mount", "backend_mount_dir",
		"web", "catalog", "views", "discovery", "repositories", "roots", "service", "telemetry",
		"logging",
	}
	var got []string
	for i := range typ.NumField() {
		if f := typ.Field(i); f.IsExported() {
			got = append(got, yamlName(f))
		}
	}
	if !slices.Equal(got, want) {
		t.Errorf("Config top-level yaml tags = %q, want %q", got, want)
	}
}

func TestContractFieldNames(t *testing.T) {
	tests := []struct {
		typ  reflect.Type
		want []string
	}{
		{
			typ:  reflect.TypeFor[Repository](),
			want: []string{"ID", "Repository", "ResticBinary", "RcloneBinary", "PasswordFile", "CacheDir", "NoCache", "LockMode", "Environment"},
		},
		{
			typ:  reflect.TypeFor[Root](),
			want: []string{"ID", "LocalPath", "RepositoryID", "PrefixMap", "Snapshots", "SeedPaths", "ExcludeRelativePaths", "Snap"},
		},
		{
			typ:  reflect.TypeFor[SeedPath](),
			want: []string{"Path", "MaxDepth"},
		},
	}
	for _, tt := range tests {
		for _, name := range tt.want {
			if _, ok := tt.typ.FieldByName(name); !ok {
				t.Errorf("%s has no field %s, want it", tt.typ.Name(), name)
			}
		}
	}

	fieldTypes := []struct {
		typ   reflect.Type
		field string
		want  reflect.Type
	}{
		{reflect.TypeFor[Repository](), "Environment", reflect.TypeFor[map[string]string]()},
		{reflect.TypeFor[SeedPath](), "Path", reflect.TypeFor[string]()},
		{reflect.TypeFor[SeedPath](), "MaxDepth", reflect.TypeFor[int]()},
	}
	for _, ft := range fieldTypes {
		f, ok := ft.typ.FieldByName(ft.field)
		if !ok {
			continue // Reported above.
		}
		if f.Type != ft.want {
			t.Errorf("%s.%s type = %v, want %v", ft.typ.Name(), ft.field, f.Type, ft.want)
		}
	}

	if got := reflect.TypeFor[Revision]().Kind(); got != reflect.String {
		t.Errorf("Revision kind = %v, want %v", got, reflect.String)
	}
}

func TestDurationFieldsAreDurations(t *testing.T) {
	tests := []struct {
		typ   reflect.Type
		field string
	}{
		{reflect.TypeFor[Catalog](), "RefreshInterval"},
		{reflect.TypeFor[Catalog](), "PresenceCacheTTL"},
		{reflect.TypeFor[OnAccess](), "HandlerTimeout"},
	}
	want := reflect.TypeFor[time.Duration]()
	for _, tt := range tests {
		f, ok := tt.typ.FieldByName(tt.field)
		if !ok {
			t.Errorf("%s has no field %s, want one of type %v", tt.typ.Name(), tt.field, want)
			continue
		}
		if f.Type != want {
			t.Errorf("%s.%s type = %v, want %v", tt.typ.Name(), tt.field, f.Type, want)
		}
	}
}
