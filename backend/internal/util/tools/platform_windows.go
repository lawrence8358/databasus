package tools

// Windows marks a program by its extension, and its file modes carry no
// executable bit to check.
const (
	executableSuffix        = ".exe"
	isExecutableBitRequired = false
)
