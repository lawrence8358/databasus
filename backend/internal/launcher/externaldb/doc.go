// Package externaldb handles the connection string of an external metadata database, which an
// administrator can choose instead of the bundled PostgreSQL. The string holds a password, so it
// is stored only in a file with a restricted ACL and never on a command line or in a log.
package externaldb
