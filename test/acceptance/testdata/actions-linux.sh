#!/bin/sh
# actions-linux.sh — the scripted minimum-action path to `ls .snapshot` on Linux.
#
# Counting rule (docs/agents/sprint5-adoption/intake.md §1): one counted action is
# one command the user types. Every counted command sits on a line ending in
# `# action`; the acceptance test counts those lines and compares the total with
# the `actions: N` line this script prints last. Environment preparation the
# harness does — creating the repository, exporting RESTIC_*, pointing XDG_* at a
# scratch home, taking the first backup of $PROJECT — is not counted. Reading
# output is not an action, and neither is waiting for a service to come up.
#
# Substitution: action 1 of the real path is
#   curl -fsSL https://snapback.run/install.sh | sh
# The installer is proven separately by the installer acceptance tests, so here it
# is represented by the one command that shows the binary is installed and on
# PATH: `snapback version`. It stands in for the install step and counts as one.
#
# Environment the harness sets: SNAPBACK_BIN, RESTIC_REPOSITORY,
# RESTIC_PASSWORD_FILE, PROJECT (a directory already backed up once) and XDG_*.
# SNAPBACK_SETUP_FLAGS is optional and defaults to empty; CI passes --no-service
# where no systemd user manager is available. The bar counts `setup` alone, so
# those flags never add an action.

set -eu

: "${SNAPBACK_BIN:?SNAPBACK_BIN must be set}"
: "${PROJECT:?PROJECT must be set}"
: "${RESTIC_REPOSITORY:?RESTIC_REPOSITORY must be set}"
: "${RESTIC_PASSWORD_FILE:?RESTIC_PASSWORD_FILE must be set}"
SNAPBACK_SETUP_FLAGS="${SNAPBACK_SETUP_FLAGS:-}"

"$SNAPBACK_BIN" version # action

# shellcheck disable=SC2086 # SNAPBACK_SETUP_FLAGS is a deliberate word list
cd "$PROJECT" && "$SNAPBACK_BIN" setup $SNAPBACK_SETUP_FLAGS # action

entries=$(ls "$PROJECT/.snapshot") # action
printf '%s\n' "$entries"
if [ -z "$entries" ]; then
	echo "actions-linux: $PROJECT/.snapshot listed no entries" >&2
	exit 1
fi

echo "actions: 3"
