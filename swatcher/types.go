package swatcher

import (
	"os"
	"os/exec"
)

const ServiceDefinitionDir = "/etc/sminit"

type Manager interface {
	Start()
	Stop()
	Monitor()
	Forget()
	Log()
	LogAll()
}

type Swatcher struct {
	Cmd            *exec.Cmd
	SortedServices []Service
}

type Status int

const (
	Running Status = iota
	Stopped
	Pending
)

type Service struct {
	Name      string
	Cmd       string
	Log       string
	RunBefore []string
	RunAfter  []string
	Stdout    Stdout
	Stderr    Stderr
}

type Stdout struct {
	File   *os.File
	Prefix string
}

type Stderr struct {
	File   *os.File
	Prefix string
}

func (s *Stdout) Write(b []byte) (n int, err error) {
	_, err = s.File.Write([]byte(s.Prefix))
	if err != nil {
		return 0, err
	}

	return s.File.Write(b)
}

func (s *Stderr) Write(b []byte) (n int, err error) {
	_, err = s.File.Write([]byte(s.Prefix))
	if err != nil {
		return 0, err
	}

	return s.File.Write(b)
}

// NewSwatcher creates a new Swatcher, it should return an error if an instance of Swatcher is already running.
func NewSwatcher() (Swatcher, error) {
	// check if there is an instance running of Swatcher
	// if there is: return an error
	// if not: create a new one
	return Swatcher{}, nil
}
