//go:build windows

// Package metadatadb runs the PostgreSQL server bundled with the Windows build, which holds
// Databasus's own metadata. Start creates the cluster on first use, repairs what an unclean
// shutdown left behind, and returns a connection string; Stop shuts the server down cleanly.
//
// The server accepts password logins from 127.0.0.1 only. The password is generated at cluster
// creation and kept in a file in the data folder, because nothing else on Windows can carry it
// between service starts the way the Docker image's Unix-socket bootstrap does.
package metadatadb
