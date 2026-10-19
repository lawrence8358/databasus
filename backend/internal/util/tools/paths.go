package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// archAssetsKey returns the assets/tools/<key>/ subdirectory name for the
// current GOOS+GOARCH. Panics on unsupported combinations because the app
// cannot operate without DB client binaries.
func archAssetsKey() string {
	switch {
	case runtime.GOOS == "linux" && runtime.GOARCH == "arm64":
		return "arm"
	case runtime.GOOS == "linux":
		return "x64"
	case runtime.GOOS == "windows" && runtime.GOARCH == "amd64":
		return "win-x64"
	}

	panic(fmt.Sprintf("unsupported OS/arch for DB client tools: %s/%s",
		runtime.GOOS, runtime.GOARCH))
}

// AssetsToolsDir returns the absolute path to assets/tools/<arch-key>/.
// Walks up from the executable's folder, then from cwd, looking for the
// directory. A Windows service starts in an arbitrary working directory, so
// the executable's own location comes first; `go run` and `go test` build into
// a temporary folder, so a development checkout is found from cwd. In Docker
// (cwd=/app) this resolves to /app/assets/tools/<arch-key>/.
var AssetsToolsDir = sync.OnceValue(func() string {
	key := archAssetsKey()
	relativePath := filepath.Join("assets", "tools", key)

	searchRoots := make([]string, 0, 2)

	if executablePath, err := os.Executable(); err == nil {
		if resolvedPath, err := filepath.EvalSymlinks(executablePath); err == nil {
			executablePath = resolvedPath
		}

		searchRoots = append(searchRoots, filepath.Dir(executablePath))
	}

	if cwd, err := os.Getwd(); err == nil {
		searchRoots = append(searchRoots, cwd)
	}

	for _, searchRoot := range searchRoots {
		if toolsDir, isFound := findFolderUpwards(searchRoot, relativePath); isFound {
			return toolsDir
		}
	}

	panic(fmt.Sprintf("could not locate %s starting from %s",
		relativePath, strings.Join(searchRoots, " or ")))
})

func findFolderUpwards(startFolder, relativePath string) (string, bool) {
	candidate := startFolder

	for {
		path := filepath.Join(candidate, relativePath)
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return path, true
		}

		parent := filepath.Dir(candidate)
		if parent == candidate {
			return "", false
		}

		candidate = parent
	}
}
