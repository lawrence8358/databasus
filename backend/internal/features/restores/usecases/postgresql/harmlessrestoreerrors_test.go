package usecases_postgresql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const publicSchemaClashStderr = `pg_restore: error: could not execute query: ERROR:  schema "public" already exists
Command was: CREATE SCHEMA public;


pg_restore: warning: errors ignored on restore: 1
`

func Test_IsEveryIgnoredRestoreErrorHarmless_WhenOnlyPublicSchemaError_ReturnsTrue(t *testing.T) {
	assert.True(t, IsEveryIgnoredRestoreErrorHarmless(publicSchemaClashStderr))
}

func Test_IsEveryIgnoredRestoreErrorHarmless_WhenMissingExtensionAlsoIgnored_ReturnsFalse(t *testing.T) {
	mixed := `pg_restore: error: could not execute query: ERROR:  schema "public" already exists
Command was: CREATE SCHEMA public;
pg_restore: error: could not execute query: ERROR:  extension "postgis" is not available
Command was: CREATE EXTENSION IF NOT EXISTS postgis WITH SCHEMA public;
pg_restore: warning: errors ignored on restore: 2`

	assert.False(t, IsEveryIgnoredRestoreErrorHarmless(mixed),
		"a user-facing restore that silently dropped an extension is still a failed restore")
}

func Test_IsEveryIgnoredRestoreErrorHarmless_WhenVisibleErrorCountBelowMarker_ReturnsFalse(t *testing.T) {
	truncated := `pg_restore: error: could not execute query: ERROR:  schema "public" already exists
Command was: CREATE SCHEMA public;
pg_restore: warning: errors ignored on restore: 4`

	assert.False(t, IsEveryIgnoredRestoreErrorHarmless(truncated),
		"stderr that cannot account for all N errors must not be tolerated")
}

func Test_IsEveryIgnoredRestoreErrorHarmless_WhenNoIgnoredErrorsMarker_ReturnsFalse(t *testing.T) {
	aborted := `pg_restore: error: could not execute query: ERROR:  schema "public" already exists
Command was: CREATE SCHEMA public;
pg_restore: error: aborting because of errors`

	assert.False(t, IsEveryIgnoredRestoreErrorHarmless(aborted))
}

func Test_IsEveryIgnoredRestoreErrorHarmless_WhenEmpty_ReturnsFalse(t *testing.T) {
	assert.False(t, IsEveryIgnoredRestoreErrorHarmless(""))
}

func Test_IsEveryIgnoredRestoreErrorHarmless_WhenRelationAlreadyExists_ReturnsFalse(t *testing.T) {
	relationClash := `pg_restore: error: could not execute query: ERROR:  relation "public.t_a" already exists
Command was: CREATE TABLE public.t_a (id integer NOT NULL);
pg_restore: warning: errors ignored on restore: 1`

	assert.False(t, IsEveryIgnoredRestoreErrorHarmless(relationClash))
}

func Test_IsEveryIgnoredRestoreErrorHarmless_WhenNewerClientSetsTransactionTimeoutOnOlderServer_ReturnsTrue(
	t *testing.T,
) {
	// pg_restore.exe ends its lines with CRLF.
	newerClientStderr := "pg_restore: connecting to database for restore\r\n" +
		"pg_restore: executing SELECT pg_catalog.set_config('search_path', '', false);\r\n" +
		"pg_restore: while INITIALIZING:\r\n" +
		"pg_restore: error: could not execute query: ERROR:  unrecognized configuration parameter \"transaction_timeout\"\r\n" +
		"Command was: SET transaction_timeout = 0;\r\n" +
		"pg_restore: creating TABLE \"public.orders\"\r\n" +
		"pg_restore: warning: errors ignored on restore: 1\r\n"

	assert.True(t, IsEveryIgnoredRestoreErrorHarmless(newerClientStderr))
}

func Test_IsEveryIgnoredRestoreErrorHarmless_WhenOtherParameterIsUnrecognized_ReturnsFalse(t *testing.T) {
	unknownParameter := `pg_restore: error: could not execute query: ERROR:  unrecognized configuration parameter "default_toast_compression"
Command was: SET default_toast_compression = 'lz4';
pg_restore: warning: errors ignored on restore: 1`

	assert.False(t, IsEveryIgnoredRestoreErrorHarmless(unknownParameter))
}

func Test_IsEveryIgnoredRestoreErrorHarmless_WhenPublicSchemaAndTransactionTimeoutBothIgnored_ReturnsTrue(
	t *testing.T,
) {
	bothHarmless := `pg_restore: error: could not execute query: ERROR:  unrecognized configuration parameter "transaction_timeout"
Command was: SET transaction_timeout = 0;
pg_restore: error: could not execute query: ERROR:  schema "public" already exists
Command was: CREATE SCHEMA public;
pg_restore: warning: errors ignored on restore: 2`

	assert.True(t, IsEveryIgnoredRestoreErrorHarmless(bothHarmless))
}
