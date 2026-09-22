// Command stage1latency measures Stage 1 restore latency against a Restic repository on an rclone remote.
package main

import "os"

func main() { os.Exit(run(os.Args[1:], os.Getenv, os.Stdout, os.Stderr)) }
