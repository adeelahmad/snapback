// Package daemon runs the Snapback daemon: it takes the single-instance lock,
// opens the IPC socket, recovers from a crashed run, starts the repository
// mounts, refreshes the catalog, then starts discovery and prewarming.
package daemon
