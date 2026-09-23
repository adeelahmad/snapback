package config

// The `telemetry:` section's fields live on the Telemetry struct in
// types.go. This file is the compile shim S6-05/T1 (RED) reserves for the
// section's validation, added once GREEN fixes the Endpoint yaml tag (see
// the TODO on Telemetry.Endpoint in types.go).
