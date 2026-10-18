package usecases_postgresql

import (
	"slices"
	"strconv"
	"strings"
)

const ignoredErrorsMarker = "errors ignored on restore:"

// pg_restore runs without --exit-on-error, so it reports "errors ignored on restore: N" and exits
// non-zero even when every ignored error left the data intact. Tolerating that is only safe when
// every ignored error is accounted for and each one is listed here with the reason it is harmless.
var harmlessRestoreErrors = []string{
	// A backup taken with include-schemas selects public explicitly, so pg_dump emits CREATE SCHEMA
	// public, and every restore target already owns that schema. A --clean restore never reaches
	// here, since the DROP removes the collision first.
	`schema "public" already exists`,

	// pg_restore 17 and later disables transaction_timeout at the start of its session. A server
	// before 17 has no such timeout, so there is nothing to disable. A client newer than its target
	// hits this on every restore, which is the normal case on Windows, where one newest client
	// serves every server version.
	`unrecognized configuration parameter "transaction_timeout"`,
}

func IsEveryIgnoredRestoreErrorHarmless(stderr string) bool {
	ignoredCount, hasMarker := parseIgnoredErrorCount(stderr)
	if !hasMarker || ignoredCount < 1 {
		return false
	}

	itemErrors := queryErrorLines(stderr)
	if len(itemErrors) != ignoredCount {
		return false
	}

	for _, line := range itemErrors {
		if !isHarmlessRestoreError(line) {
			return false
		}
	}

	return true
}

func isHarmlessRestoreError(errorLine string) bool {
	lowerErrorLine := strings.ToLower(errorLine)

	return slices.ContainsFunc(harmlessRestoreErrors, func(harmlessError string) bool {
		return strings.Contains(lowerErrorLine, harmlessError)
	})
}

func parseIgnoredErrorCount(stderr string) (int, bool) {
	markerIndex := strings.LastIndex(strings.ToLower(stderr), ignoredErrorsMarker)
	if markerIndex < 0 {
		return 0, false
	}

	fields := strings.Fields(stderr[markerIndex+len(ignoredErrorsMarker):])
	if len(fields) == 0 {
		return 0, false
	}

	count, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0, false
	}

	return count, true
}

func queryErrorLines(stderr string) []string {
	var errorLines []string

	for line := range strings.SplitSeq(stderr, "\n") {
		if strings.Contains(strings.ToLower(line), "could not execute query: error:") {
			errorLines = append(errorLines, line)
		}
	}

	return errorLines
}
