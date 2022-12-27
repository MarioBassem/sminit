package swatcher

import (
	"os"

	"github.com/spf13/cobra"
)

type Manager interface {
	Start()
	Stop()
	Monitor()
	Forget()
	Log()
	LogAll()
}

type Swatcher struct {
	PID int
}

type Status int

const (
	Running Status = iota
	Stopped
	Pending
)

type Service struct {
	Status   Status
	Children []Service
	Cmd      *cobra.Command
	LogsPath string
}

func NewSwatcher() (Swatcher, error) {
	pid := os.Getpid()
	return Swatcher{pid}, nil
}

func Watch() error {
	// this is the entry point of Swatcher
	// fetch all service definition files from /etc/sminit
	// create relationships between services
	// start services
	// wait for input from another process
	// watch running service and respawn them if they were terminated.
	return nil
}
