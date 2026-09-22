package restic

import (
	"strings"
	"testing"
)

// Fake values only: nothing here is or resembles a real credential.
const (
	fakeRepo     = "s3:s3.example.invalid/fake-bucket"
	fakePwFile   = "/etc/snapback/fake-password.txt"
	fakePassword = "not-a-real-password"
	fakeAWSKey   = "FAKE-AWS-SECRET-VALUE"
	fakeB2Key    = "FAKE-B2-ACCOUNT-KEY"
	fakeRclone   = "FAKE-RCLONE-CONFIG-PASS"
	fakeToken    = "FAKE-GITHUB-TOKEN"
)

// snapshotsArgs mirrors the argv the runner builds for a snapshots listing.
func snapshotsArgs() []string {
	return []string{
		"restic", "-r", fakeRepo,
		"--password-file", fakePwFile,
		"--no-lock", "snapshots", "--json",
	}
}

// secretEnv mirrors the child env shape, sorted as childEnv sorts it.
func secretEnv() []string {
	return []string{
		"AWS_SECRET_ACCESS_KEY=" + fakeAWSKey,
		"B2_ACCOUNT_KEY=" + fakeB2Key,
		"GITHUB_TOKEN=" + fakeToken,
		"PATH=/usr/bin:/bin",
		"RCLONE_CONFIG_PASS=" + fakeRclone,
		"RESTIC_PASSWORD=" + fakePassword,
		"RESTIC_PASSWORD_FILE=" + fakePwFile,
		"RESTIC_REPOSITORY=" + fakeRepo,
	}
}

func TestCommandLine(t *testing.T) {
	tests := []struct {
		name string
		args []string
		env  []string
		want string
	}{
		{
			name: "nil args and env",
			want: "",
		},
		{
			name: "empty slices",
			args: []string{},
			env:  []string{},
			want: "",
		},
		{
			name: "argv only, no secrets",
			args: []string{"restic", "version"},
			want: "restic version",
		},
		{
			name: "password-file value redacted",
			args: []string{"restic", "--password-file", fakePwFile, "snapshots"},
			want: "restic --password-file *** snapshots",
		},
		{
			name: "password-file equals form redacted",
			args: []string{"restic", "--password-file=" + fakePwFile, "snapshots"},
			want: "restic --password-file=*** snapshots",
		},
		{
			name: "trailing password-file with no value",
			args: []string{"restic", "--password-file"},
			want: "restic --password-file",
		},
		{
			name: "non-secret env kept verbatim",
			env:  []string{"PATH=/usr/bin:/bin", "RESTIC_REPOSITORY=" + fakeRepo},
			args: []string{"restic", "snapshots"},
			want: "PATH=/usr/bin:/bin RESTIC_REPOSITORY=" + fakeRepo + " restic snapshots",
		},
		{
			name: "every secret env key redacted",
			env:  secretEnv(),
			args: snapshotsArgs(),
			want: "AWS_SECRET_ACCESS_KEY=*** B2_ACCOUNT_KEY=*** GITHUB_TOKEN=*** " +
				"PATH=/usr/bin:/bin RCLONE_CONFIG_PASS=*** RESTIC_PASSWORD=*** " +
				"RESTIC_PASSWORD_FILE=*** RESTIC_REPOSITORY=" + fakeRepo + " " +
				"restic -r " + fakeRepo + " --password-file *** --no-lock snapshots --json",
		},
		{
			name: "fields with spaces are quoted",
			args: []string{"restic", "backup", "--exclude", "/home/fake user/Movies", "--", "/home/fake user"},
			env:  []string{"RESTIC_REPOSITORY=/mnt/backup drive/repo"},
			want: `RESTIC_REPOSITORY="/mnt/backup drive/repo" restic backup --exclude ` +
				`"/home/fake user/Movies" -- "/home/fake user"`,
		},
		{
			name: "env pair without a value is kept",
			env:  []string{"RESTIC_PASSWORD"},
			args: []string{"restic"},
			want: "RESTIC_PASSWORD restic",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CommandLine(tc.args, tc.env)
			if got != tc.want {
				t.Errorf("CommandLine() =\n  %q\nwant\n  %q", got, tc.want)
			}
		})
	}
}

func TestCommandLineLeaksNoSecret(t *testing.T) {
	secrets := []string{
		fakePwFile, fakePassword, fakeAWSKey, fakeB2Key, fakeRclone, fakeToken,
	}
	got := CommandLine(snapshotsArgs(), secretEnv())
	if got == "" {
		t.Fatal("CommandLine() = \"\", want a rendered command line")
	}
	for _, s := range secrets {
		if strings.Contains(got, s) {
			t.Errorf("CommandLine() leaks %q in %q", s, got)
		}
	}
	if !strings.Contains(got, fakeRepo) {
		t.Errorf("CommandLine() = %q, want the non-secret repository %q kept verbatim", got, fakeRepo)
	}
}
