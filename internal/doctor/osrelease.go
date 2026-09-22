package doctor

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

// parseOSRelease reads an /etc/os-release body and returns its ID and ID_LIKE values,
// lowercased and unquoted. Unknown or absent keys yield empty results.
func parseOSRelease(r io.Reader) (id string, idLike []string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		key, value, ok := strings.Cut(strings.TrimSpace(scanner.Text()), "=")
		if !ok {
			continue
		}
		value = strings.ToLower(strings.Trim(strings.TrimSpace(value), `"'`))
		switch strings.TrimSpace(key) {
		case "ID":
			id = value
		case "ID_LIKE":
			idLike = strings.Fields(value)
		}
	}
	return id, idLike
}

// readOSRelease parses the os-release file at path. A missing or unreadable file
// yields empty results, because the distribution is then simply unknown.
func readOSRelease(path string) (string, []string) {
	f, err := os.Open(path)
	if err != nil {
		return "", nil
	}
	defer func() { _ = f.Close() }()
	return parseOSRelease(f)
}

// packageCommand returns the command that installs pkg on the distribution named by id
// and idLike. An unrecognised distribution yields generic advice rather than a command
// that would not work there.
func packageCommand(id string, idLike []string, pkg string) string {
	families := []struct {
		ids    []string
		likes  []string
		format string
	}{
		{
			ids:    []string{"debian", "ubuntu", "raspbian"},
			likes:  []string{"debian"},
			format: "sudo apt install %s",
		},
		{
			ids:    []string{"fedora", "rhel", "centos", "rocky", "almalinux", "alma"},
			likes:  []string{"rhel", "fedora"},
			format: "sudo dnf install %s",
		},
		{
			ids:    []string{"arch", "manjaro"},
			format: "sudo pacman -S %s",
		},
		{
			ids:    []string{"alpine"},
			format: "sudo apk add %s",
		},
		{
			likes:  []string{"suse"},
			format: "sudo zypper install %s",
		},
	}

	for _, family := range families {
		if slices.Contains(family.ids, id) {
			return fmt.Sprintf(family.format, pkg)
		}
		for _, like := range family.likes {
			if slices.Contains(idLike, like) {
				return fmt.Sprintf(family.format, pkg)
			}
		}
	}
	if strings.HasPrefix(id, "opensuse") || id == "sles" || id == "sled" {
		return fmt.Sprintf("sudo zypper install %s", pkg)
	}
	return fmt.Sprintf("install the %s package with your distribution's package manager", pkg)
}
