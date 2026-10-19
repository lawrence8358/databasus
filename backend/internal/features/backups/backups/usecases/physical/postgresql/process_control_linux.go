package usecases_physical_postgresql

import (
	"os"
	"os/exec"
	"syscall"
)

func signalForGracefulCancel(p *os.Process) error {
	if p == nil {
		return nil
	}

	return p.Signal(os.Interrupt)
}

// Pdeathsig makes the kernel SIGTERM pg_receivewal if the Databasus process
// dies, so a crashed supervisor never leaks an orphaned receiver that keeps
// holding the replication slot.
func setReceivewalProcessAttributes(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGTERM}
}
