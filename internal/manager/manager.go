package manager

import (
	"context"
	"strings"
	"sync"

	"fmt"
	"os/exec"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"
)

var (
	ErrSminitInternalError = errors.New("sminit internal error")
	ErrBadRequest          = errors.New("bad request")
)

// Manager handles service manipulation
type Manager struct {
	services map[string]*Service
	mut      sync.RWMutex
}

type stdoutLogger struct {
	serviceName string
}
type stderrLogger struct {
	serviceName string
}

// Status presents service status
type Status string

const (
	// service received a start signal
	Started Status = "started"
	// service is running and healthy
	Running Status = "running"
	// service completed its task and process terminated with exit status 0
	Successful Status = "successful"
	// service did not complete its task and process terminated with exit status other than 0
	Failed Status = "failed"
	// service is waiting its pending parents to run
	Pending Status = "pending"
	// service is stopped by user, should not be started unless user requested
	Stopped Status = "stopped"
)

// Service contains all needed information during the lifetime of a service
type Service struct {
	Name   string
	Status Status

	// children are services that depend on this service.
	children map[string]bool
	// parents are services that this service depend on.
	parents     map[string]bool
	log         string
	healthCheck string
	oneShot     bool
	cmdStr      string
	stdout      stdoutLogger
	stderr      stderrLogger

	startSignal  chan bool
	deleteSignal chan bool
	stopSignal   chan bool

	isStopped chan bool
	isDeleted chan bool
	mut       sync.RWMutex
}

// NewManager creates a new Manager struct and populates it with services generated from provided serviceOptions
func NewManager(serviceOptions map[string]ServiceOptions) (*Manager, error) {
	manager := Manager{
		services: make(map[string]*Service),
	}

	err := manager.populateServices(serviceOptions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to populate manager with services")
	}

	return &manager, nil
}

func (m *Manager) getService(name string) (*Service, bool) {
	m.mut.RLock()
	defer m.mut.RUnlock()
	s, ok := m.services[name]
	return s, ok
}

func (m *Manager) getServicesMap() map[string]*Service {
	m.mut.RLock()
	defer m.mut.RUnlock()
	return m.services
}

func (m *Manager) addService(service *Service) error {
	m.mut.Lock()
	defer m.mut.Unlock()

	if _, ok := m.services[service.Name]; ok {
		return fmt.Errorf("service with the same name (%s) exists", service.Name)
	}
	m.services[service.Name] = service
	return nil
}

func (m *Manager) deleteService(name string) {
	m.mut.Lock()
	defer m.mut.Unlock()
	delete(m.services, name)
}

func (m *Manager) populateServices(serviceOptions map[string]ServiceOptions) error {
	for _, opts := range serviceOptions {
		newService := newService(opts)
		m.addService(newService)
	}

	for name, opts := range serviceOptions {
		err := m.addToGraph(opts)
		if err != nil {
			return errors.Wrapf(err, "failed to add service %s to graph", name)
		}
	}

	return nil
}

// fireServices is responsible for starting a go routine for each service, and starting it if eligible
func (m *Manager) fireServices() {
	for name, service := range m.services {
		go m.serviceRoutine(name)
		if m.isEligibleToRun(name) {
			service.startSignal <- true
		}
	}
}

// Add adds a new service to the list of services tracked by the manager
func (m *Manager) Add(opts ServiceOptions) error {
	// generate Service struct
	// fire go routine for service
	// check if all parents are in Running or Successful state, if true, send a start signal for this service
	// return

	if _, ok := m.getService(opts.Name); ok {
		return errors.Wrapf(ErrBadRequest, "failed to add %s. a service with the same name is already tracked", opts.Name)
	}

	service := newService(opts)
	err := m.addService(service)
	if err != nil {
		return errors.Wrapf(ErrBadRequest, "failed to add service. %s", err.Error())
	}

	err = m.addToGraph(opts)
	if err != nil {
		m.deleteService(opts.Name)
		return errors.Wrapf(ErrBadRequest, "failed to modify service graph. %s", err.Error())
	}

	go m.serviceRoutine(opts.Name)

	if m.isEligibleToRun(opts.Name) {
		service.startSignal <- true
	}

	return nil
}

// Delete deletes a services with the given name from the list of services tracked by the manager
func (m *Manager) Delete(name string) error {
	// cancel service context
	// send delete signal
	// remove from service map, and from graph
	// return
	service, ok := m.getService(name)

	if !ok {
		return errors.Wrapf(ErrBadRequest, "there is no tracked service with name %s", name)
	}

	service.deleteSignal <- true
	<-service.isDeleted
	m.deleteService(name)

	SminitLog.Info().Msgf("service %s is deleted", name)
	return nil
}

// Start starts a services that is already tracked by the manager.
func (m *Manager) Start(name string) error {
	// check parents' statuses of service
	// if all are running or successful, send start signal
	// return
	service, ok := m.getService(name)

	if !ok {
		return errors.Wrapf(ErrBadRequest, "there is no tracked service with name %s", name)
	}

	if service.hasStarted() {
		return errors.Wrapf(ErrBadRequest, "service %s status is %s", name, service.Status)
	}

	service.changeStatus(Pending)

	if !m.isEligibleToRun(name) {
		return errors.Wrapf(ErrBadRequest, "service %s is still pending", name)
	}

	service.startSignal <- true

	return nil
}

// Stop stops a service that is already tracked by the manager.
func (m *Manager) Stop(name string) error {
	// cancel service context.
	// return
	service, ok := m.getService(name)

	if !ok {
		return errors.Wrapf(ErrBadRequest, "there is no tracked service with name %s", name)
	}

	service.stopSignal <- true
	<-service.isStopped

	return nil
}

// List lists all services tracked by the manager.
func (m *Manager) List() []Service {
	// list all services with their statuses
	// return
	services := m.getServicesMap()

	var ret []Service

	for _, service := range services {
		ret = append(ret, *service)
	}
	return ret
}

func (m *Manager) serviceRoutine(name string) {
	service, ok := m.getService(name)
	if !ok {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	for {
		select {
		case <-service.startSignal:
			go m.runService(ctx, name)

		case <-service.stopSignal:
			cancel()

		case <-service.deleteSignal:
			if !service.hasStarted() {
				service.isDeleted <- true
				return
			}
			cancel()
			<-service.isStopped
			service.isDeleted <- true
			return
		}
	}
}

func (m *Manager) runService(ctx context.Context, serviceName string) {
	service, ok := m.getService(serviceName)
	if !ok {
		return
	}

	// service status is started
	service.changeStatus(Started)

	err := backoff.Retry(func() error {
		select {
		case <-ctx.Done():
			service.changeStatus(Stopped)
			service.isStopped <- true
			return backoff.Permanent(fmt.Errorf("service %s was stopped", service.Name))

		default:
			splittedCmd := strings.Split(service.cmdStr, " ")
			cmd := exec.CommandContext(ctx, splittedCmd[0], splittedCmd[1:]...)
			if service.log == "stdout" {
				cmd.Stdout = &service.stdout
				cmd.Stderr = &service.stderr
			}

			err := cmd.Start()
			if err != nil {
				SminitLog.Error().Msgf("error while starting process %s. %s", service.Name, err.Error())
				return errors.New("restarting service")
			}

			if !isHealthy(ctx, service) {
				err = cmd.Process.Kill()
				if err != nil {
					SminitLog.Error().Msgf("error killing process %s. %s", service.Name, err.Error())
				}
				return errors.New("service is not healthy. restarting...")
			}

			// service status is running
			service.changeStatus(Running)

			m.startEligibleChildren(service.Name)

			err = cmd.Wait()
			if err != nil {
				service.changeStatus(Failed)
				SminitLog.Error().Msgf("error while running process %s. %s", service.Name, err.Error())

				return errors.New("restarting service")
			}

			service.changeStatus(Successful)
			if service.oneShot {
				return backoff.Permanent(fmt.Errorf("service %s has finished", service.Name))
			}
			time.Sleep(100 * time.Millisecond)

			return errors.New("restarting service")
		}
	}, newExponentialBackOff())

	SminitLog.Info().Msg(err.Error())

	if service.oneShot {
		return
	}

}

// this function will return false only if context was cancelled or backoff timesout, and true if cmd.Run() returned nil, i.e process is healthy
func isHealthy(ctx context.Context, service *Service) bool {
	exponentialBackoff := newExponentialBackOff()
	exponentialBackoff.MaxElapsedTime = time.Minute
	healthy := false
	err := backoff.Retry(func() error {
		select {
		case <-ctx.Done():
			return backoff.Permanent(errors.New("context canceled"))
		default:
			splittedCmd := strings.Split(service.healthCheck, " ")
			cmd := exec.CommandContext(ctx, splittedCmd[0], splittedCmd[1:]...)
			err := cmd.Run()
			if err == nil {
				healthy = true
				return backoff.Permanent(errors.New("health check is successful"))
			}
			return errors.New("health check failed")
		}
	}, exponentialBackoff)
	SminitLog.Trace().Msgf("service %s health check: %s", service.Name, err.Error())
	return healthy

}

func (m *Manager) startEligibleChildren(name string) {
	// check if dependent services are eligible to be run
	service, _ := m.getService(name)

	service.mut.RLock()
	defer service.mut.RUnlock()

	for childName := range service.children {
		child, ok := m.getService(childName)
		if !ok {
			continue
		}
		if child.hasStarted() || !m.isEligibleToRun(childName) {
			continue
		}

		child.startSignal <- true

	}
}

func (s *Service) hasStarted() bool {
	s.mut.RLock()
	defer s.mut.RUnlock()

	return s.Status == Running || s.Status == Started || s.Status == Successful || s.Status == Failed
}

func (m *Manager) isEligibleToRun(name string) bool {
	// a service is said to be eligible to run if it is pending, and all its parents are running or successful (healthy)
	service, _ := m.getService(name)

	service.mut.RLock()
	defer service.mut.RUnlock()

	if service.Status != Pending {
		return false
	}

	for parentName := range service.parents {
		parent, ok := m.getService(parentName)
		if !ok || !parent.isRunningOrSuccessful() {
			return false
		}
	}

	return true
}

func (s *Service) isRunningOrSuccessful() bool {
	s.mut.RLock()
	defer s.mut.RUnlock()
	return s.Status == Running || s.Status == Successful

}

func (s *Service) changeStatus(newStatus Status) {
	s.mut.Lock()
	s.Status = newStatus
	s.mut.Unlock()
}

func newService(service ServiceOptions) *Service {
	stdout := stdoutLogger{
		serviceName: service.Name,
	}
	stderr := stderrLogger{
		serviceName: service.Name,
	}

	healthCheck := service.HealthCheck
	if healthCheck == "" {
		healthCheck = "sleep 1"
	}

	newService := Service{
		Name:         service.Name,
		Status:       Pending,
		log:          service.Log,
		healthCheck:  healthCheck,
		cmdStr:       service.Cmd,
		oneShot:      service.OneShot,
		stdout:       stdout,
		stderr:       stderr,
		children:     map[string]bool{},
		parents:      map[string]bool{},
		startSignal:  make(chan bool),
		stopSignal:   make(chan bool),
		deleteSignal: make(chan bool),
		isStopped:    make(chan bool),
		isDeleted:    make(chan bool),
		mut:          sync.RWMutex{},
	}
	return &newService
}

func (m *Manager) addToGraph(service ServiceOptions) error {
	for _, parent := range service.After {
		if _, ok := m.getService(parent); !ok {
			return errors.Wrapf(ErrBadRequest, "service %s does not exist", parent)
		}
		m.services[service.Name].parents[parent] = true
		m.services[parent].children[service.Name] = true
	}
	return nil
}

func newExponentialBackOff() *backoff.ExponentialBackOff {
	b := backoff.ExponentialBackOff{
		InitialInterval:     backoff.DefaultInitialInterval,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
		Multiplier:          backoff.DefaultMultiplier,
		MaxInterval:         1 * time.Second,
		MaxElapsedTime:      0,
		Clock:               backoff.SystemClock,
	}
	b.Reset()
	return &b
}

func (l *stderrLogger) Write(p []byte) (int, error) {
	SminitLog.Error().Str("component", fmt.Sprintf("%s:", l.serviceName)).Msg(string(p[:len(p)-1]))
	return len(p), nil
}

func (l *stdoutLogger) Write(p []byte) (int, error) {
	SminitLog.Info().Str("component", fmt.Sprintf("%s:", l.serviceName)).Msg(string(p[:len(p)-1]))
	return len(p), nil
}

/*
	manager is responsible for manipulating services
	one instance of the manager should be acquired by the server

	service maintains memory state
	for each service, there are two routines running:
		1- service routine which runs the actual process of the service
		2- signal routine which is responsible for starting/stopping the process at any given time

	service shared memory:
		beside channels, a service has some memory shared between go routines:
			1- service status
			2- service parents
			3- service children
		there should be locks before dealing with any of these states

	when should a service receive a start signal:
		a service receives a start signal when all of its parents are not pending and tracked
		a check could be initiated by two factors:
			1- a user wants to start the service
			2- a parent service passed its health check

	what should happen when deleting a service?
		1- service should be stopped, deleted from manager's tracked services.
		2- service graph should not be modified

*/
