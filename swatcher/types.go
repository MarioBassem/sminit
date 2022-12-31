package swatcher

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
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
	Cmd                 *exec.Cmd
	UserDefinedServices []Service
	SortedServices      []Service
}

type Status int

const (
	Running Status = iota
	Stopped
	Pending
)

type Service struct {
	Name      string
	CmdStr    string
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
	str := strings.Join([]string{s.Prefix, string(b), "\n"}, "")
	return s.File.Write([]byte(str))
}

func (s *Stderr) Write(b []byte) (n int, err error) {
	str := strings.Join([]string{s.Prefix, string(b), "\n"}, "")
	return s.File.Write([]byte(str))
}

func NewSwatcher() (Swatcher, error) {
	return Swatcher{}, nil
}

func Watch() error {
	// this is the entry point of Swatcher
	// fetch all service definition files from /etc/sminit
	// create relationships between services
	// start services
	// wait for input from another process
	// watch running service and respawn them if they were terminated.
	swatcher, err := NewSwatcher()
	if err != nil {
		return errors.Wrapf(err, "couldn't create a new Swatcher")
	}

	userDefinedServices, err := swatcher.LoadAll(ServiceDefinitionDir)
	if err != nil {
		return errors.Wrapf(err, "couldn't load services")
	}

	swatcher.UserDefinedServices = userDefinedServices

	// services, err = Sort(userDefinedServices)
	// swatcher.SortedServices = services
	// swatcher.Watch()
	// swatcher.Sort()
	return nil
}

func NewService(service Service) Service {

	return Service{
		Name:   service.Name,
		CmdStr: service.CmdStr,
		Stdout: Stdout{
			File:   os.Stdout,
			Prefix: fmt.Sprintf("[+]%s: ", service.Name),
		},
		Stderr: Stderr{
			File:   os.Stderr,
			Prefix: fmt.Sprintf("[-]%s: ", service.Name),
		},
	}
}
