package usecases_physical_postgresql

import (
	"os"
	"os/exec"
)

// Windows cannot deliver an interrupt to a child process, so cancellation
// ends it outright. Physical backups are unavailable on Windows, so this only
// has to keep the package compiling and behaving sanely.
func signalForGracefulCancel(p *os.Process) error {
	if p == nil {
		return nil
	}

	return p.Kill()
}

func setReceivewalProcessAttributes(_ *exec.Cmd) {}
