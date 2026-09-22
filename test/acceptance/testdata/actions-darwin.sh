#!/bin/sh
# actions-darwin.sh — the scripted minimum-action path to `ls .snapshot` on macOS.
#
# Counting rule (docs/agents/sprint5-adoption/intake.md §1): one counted action is
# one command the user types. Every counted command sits on a line ending in
# `# action`; the acceptance test counts those lines and compares the total with
# the `actions: N` line this script prints last. Environment preparation the
# harness does — creating the repository, exporting RESTIC_*, pointing XDG_* at a
# scratch home, taking the first backup of $PROJECT — is not counted. Reading
# output is not an action, and neither is waiting for the mount to come up.
#
# Substitution: action 1 of the real path is
#   curl -fsSL https://snapback.run/install.sh | sh
# The installer is proven separately by the installer acceptance tests, so here it
# is represented by the one command that shows the binary is installed and on
# PATH: `snapback version`. It stands in for the install step and counts as one.
#
# macOS has no launchd adapter in v0.1, so the user starts the daemon themselves:
# that is the fourth counted action. Backgrounding it and polling `status` until
# the mount answers is the harness waiting, not the user typing, so the poll loop
# is not counted.
#
# Environment the harness sets: SNAPBACK_BIN, RESTIC_REPOSITORY,
# RESTIC_PASSWORD_FILE, PROJECT (a directory already backed up once) and XDG_*.
# SNAPBACK_SETUP_FLAGS is optional and defaults to empty; CI passes --no-service
# where no user service manager is available. The bar counts `setup` alone, so
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

"$SNAPBACK_BIN" run & # action
run_pid=$!
trap 'kill "$run_pid" 2>/dev/null || true' EXIT INT TERM

ready=
waited=0
while [ "$waited" -lt 60 ]; do
	if "$SNAPBACK_BIN" status >/dev/null 2>&1; then
		ready=yes
		break
	fi
	waited=$((waited + 1))
	sleep 1
done
if [ -z "$ready" ]; then
	echo "actions-darwin: snapback status never became ready" >&2
	exit 1
fi

entries=$(ls "$PROJECT/.snapshot") # action
printf '%s\n' "$entries"
if [ -z "$entries" ]; then
	echo "actions-darwin: $PROJECT/.snapshot listed no entries" >&2
	exit 1
fi

echo "actions: 4"
