// Package config defines all compile-time constants and runtime configuration
// for the tunnel server and client.
//
// Concurrency model: configuration is constructed during process startup and
// then treated as immutable; no locks are required for reads after init.
// Data flow: main packages parse flags/env into Config and pass it down to
// sub-packages during initialization.
package config
