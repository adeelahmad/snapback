package setup

import "testing"

func TestApplyEnv(t *testing.T) {
	fromMap := func(env map[string]string) func(string) string {
		return func(key string) string { return env[key] }
	}

	tests := []struct {
		name   string
		start  Result
		getenv func(string) string
		want   Result
	}{
		{
			name: "both set",
			getenv: fromMap(map[string]string{
				"RESTIC_REPOSITORY":    "/srv/restic",
				"RESTIC_PASSWORD_FILE": "/etc/restic.pass",
			}),
			want: Result{RepoURI: "/srv/restic", CredentialFile: "/etc/restic.pass"},
		},
		{
			name: "surrounding whitespace trimmed",
			getenv: fromMap(map[string]string{
				"RESTIC_REPOSITORY":    "  /srv/restic\n",
				"RESTIC_PASSWORD_FILE": "\t/etc/restic.pass  ",
			}),
			want: Result{RepoURI: "/srv/restic", CredentialFile: "/etc/restic.pass"},
		},
		{
			name:  "whitespace-only counts as unset",
			start: Result{RepoURI: "kept", CredentialFile: "kept too"},
			getenv: fromMap(map[string]string{
				"RESTIC_REPOSITORY":    "   ",
				"RESTIC_PASSWORD_FILE": "\t\n",
			}),
			want: Result{RepoURI: "kept", CredentialFile: "kept too"},
		},
		{
			name:   "unset env leaves fields untouched",
			start:  Result{RepoURI: "kept", CredentialFile: "kept too"},
			getenv: fromMap(nil),
			want:   Result{RepoURI: "kept", CredentialFile: "kept too"},
		},
		{
			name:  "only repository set",
			start: Result{CredentialFile: "kept too"},
			getenv: fromMap(map[string]string{
				"RESTIC_REPOSITORY": "/srv/restic",
			}),
			want: Result{RepoURI: "/srv/restic", CredentialFile: "kept too"},
		},
		{
			name:   "nil getenv is a no-op",
			start:  Result{RepoURI: "kept", CredentialFile: "kept too"},
			getenv: nil,
			want:   Result{RepoURI: "kept", CredentialFile: "kept too"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.start
			applyEnv(&got, test.getenv)

			if got.RepoURI != test.want.RepoURI {
				t.Errorf("applyEnv(%s).RepoURI = %q, want %q", test.name, got.RepoURI, test.want.RepoURI)
			}
			if got.CredentialFile != test.want.CredentialFile {
				t.Errorf("applyEnv(%s).CredentialFile = %q, want %q", test.name, got.CredentialFile, test.want.CredentialFile)
			}
		})
	}
}
