//go:build windows

// Package launcher does on Windows what the Docker image's entrypoint script does on Linux: it
// starts the bundled metadata database, writes the Web UI's runtime configuration, runs the
// backend with the metadata connection string, and on a stop request shuts the backend down
// before the database.
package launcher
