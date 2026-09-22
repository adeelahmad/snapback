package setup

import "testing"

func TestNextAfterSetup(t *testing.T) {
	tests := []struct {
		name             string
		goos             string
		serviceInstalled bool
		daemonRunning    bool
		linked           []string
		advice           Advice
		want             string
	}{
		{
			name:   "empty repository keeps the snap advice",
			goos:   "linux",
			linked: []string{"/home/a/work"},
			advice: Advice{Next: "snapback snap /home/a/work"},
			want:   "snapback snap /home/a/work",
		},
		{
			name:             "snap advice wins over an installed service",
			goos:             "linux",
			serviceInstalled: true,
			linked:           []string{"/home/a/work"},
			advice:           Advice{Next: "snapback snap /home/a/work"},
			want:             "snapback snap /home/a/work",
		},
		{
			name:             "linux with the service installed browses the first root",
			goos:             "linux",
			serviceInstalled: true,
			linked:           []string{"/home/a/work", "/home/a/docs"},
			advice:           Advice{Next: "snapback run"},
			want:             "ls /home/a/work/.snapshot",
		},
		{
			name:   "linux without the service asks for the daemon",
			goos:   "linux",
			linked: []string{"/home/a/work"},
			advice: Advice{Next: "snapback run"},
			want:   "snapback run",
		},
		{
			name:             "darwin asks for the daemon even with a service",
			goos:             "darwin",
			serviceInstalled: true,
			linked:           []string{"/Users/a/work"},
			advice:           Advice{Next: "snapback run"},
			want:             "snapback run",
		},
		{
			name:          "a running daemon browses the first root",
			goos:          "darwin",
			daemonRunning: true,
			linked:        []string{"/Users/a/work"},
			advice:        Advice{Next: "snapback run"},
			want:          "ls /Users/a/work/.snapshot",
		},
		{
			name:             "nothing linked asks for the daemon",
			goos:             "linux",
			serviceInstalled: true,
			advice:           Advice{Next: "snapback run"},
			want:             "snapback run",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NextAfterSetup(tt.goos, tt.serviceInstalled, tt.daemonRunning, tt.linked, tt.advice)
			if got != tt.want {
				t.Errorf("NextAfterSetup(%q, %t, %t, %v, %+v) = %q, want %q",
					tt.goos, tt.serviceInstalled, tt.daemonRunning, tt.linked, tt.advice, got, tt.want)
			}
		})
	}
}
