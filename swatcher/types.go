package swatcher

import (
	"os/exec"

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
	UserDefinedServices []UserDefinedService
	SortedServices      []Service
}

type Status int

const (
	Running Status = iota
	Stopped
	Pending
)

type Service struct {
	Name   string
	CmdStr string
	Logs   string
}

type UserDefinedService struct {
	Name      string
	Cmd       string
	RunBefore []string
	RunAfter  []string
	Log       string
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

func NewService(userDefinedService UserDefinedService) Service {

	return Service{
		Name:   userDefinedService.Name,
		CmdStr: userDefinedService.Cmd,
	}
}
