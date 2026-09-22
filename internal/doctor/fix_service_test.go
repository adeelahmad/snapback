package doctor

import "testing"

func TestDaemonFixText(t *testing.T) {
	tests := []struct {
		name             string
		goos             string
		managerSupported bool
		want             string
	}{
		{
			name:             "linux with a supported manager installs the service",
			goos:             "linux",
			managerSupported: true,
			want:             "start it with snapback install service",
		},
		{
			name:             "linux without a supported manager runs in the foreground",
			goos:             "linux",
			managerSupported: false,
			want:             "start it with snapback run",
		},
		{
			name:             "darwin names the Linux-only limit",
			goos:             "darwin",
			managerSupported: false,
			want:             "start it with snapback run (the login service is Linux-only for now)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := daemonFixText(tt.goos, tt.managerSupported)
			if got != tt.want {
				t.Errorf("daemonFixText(%q, %t) = %q, want %q", tt.goos, tt.managerSupported, got, tt.want)
			}
		})
	}
}
