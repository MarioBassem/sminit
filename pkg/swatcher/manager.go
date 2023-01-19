package swatch

import (
	"context"
	"io"
	"path"
	"strings"
	"sync"

	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	ErrSminitInternalError = errors.New("sminit internal error")
	ErrBadRequest          = errors.New("bad request")
)

// Manager handles service manipulation
type Manager struct {
	services map[string]*Service
	sync.Mutex
}

// Status presents service status
type Status string

const (
	Running    Status = "running"
	Successful Status = "successful"
	Started    Status = "started"

	Pending Status = "pending"
	Failure Status = "failure"
	Stopped Status = "stopped"
)

// Service contains all needed information during the lifetime of a service
type Service struct {
	Name   string
	Status Status

	startSignal  chan bool
	deleteSignal chan bool
	stopSignal   chan bool
	// children are services that depend on this service.
	children map[string]bool
	// parents are services that this service depend on.
	parents      map[string]bool
	log          string
	healthCheck  string
	oneShot      bool
	cmdStr       string
	stdoutLogger io.Writer
	stderrLogger io.Writer

	// isHealthy indicates that a process have passed at least one health check, and until now have not exited with an error
	isHealthy bool
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

func (m *Manager) populateServices(serviceOptions map[string]ServiceOptions) error {
	for name, opts := range serviceOptions {
		newService := generateService(opts)
		m.services[name] = newService
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
	for name := range m.services {
		go m.serviceRoutine(name)
		m.startIfEligible(name)
	}
}

// Add adds a new service to the list of services tracked by the manager
func (m *Manager) Add(serviceName string) error {
	// generate Service struct
	// fire go routine for service
	// check if all parents are in Running or Successful state, if true, send a start signal for this service
	// return
	m.Lock()
	defer m.Unlock()

	if _, ok := m.services[serviceName]; ok {
		return errors.Wrapf(ErrBadRequest, "failed to add %s. a service with the same name is already tracked", serviceName)
	}

	fileName := strings.Join([]string{serviceName, ".yaml"}, "")
	path := path.Join(ServiceDefinitionDir, fileName)
	serviceOptions, err := Load(path, serviceName)
	if err != nil {
		return errors.Wrapf(err, "could not load service %s", serviceName)
	}

	newService := generateService(serviceOptions)
	m.services[newService.Name] = newService
	err = m.addToGraph(serviceOptions)
	if err != nil {
		return errors.Wrap(err, "failed to modify service graph")
	}

	go m.serviceRoutine(newService.Name)
	m.startIfEligible(newService.Name)

	SminitLog.Info().Msgf("service %s is added", serviceName)

	return nil
}

// Delete deletes a services with the given name from the list of services tracked by the manager
func (m *Manager) Delete(name string) error {
	// cancel service context
	// send delete signal
	// remove from service map, and from graph
	// return
	m.Lock()
	defer m.Unlock()

	if _, ok := m.services[name]; !ok {
		return errors.Wrapf(ErrBadRequest, "there is no tracked service with name %s", name)
	}

	service := m.services[name]
	service.stopSignal <- true
	service.deleteSignal <- true
	// for parent := range service.parents {
	// 	delete(m.services[parent].children, name)
	// }
	// for child := range service.children {
	// 	delete(m.services[child].parents, name)
	// 	m.startIfEligible(child)
	// }
	delete(m.services, name)
	SminitLog.Info().Msgf("service %s is deleted", name)
	return nil
}

// Start starts a services that is already tracked by the manager.
func (m *Manager) Start(name string) error {
	// check parents' statuses of service
	// if all are running or successful, send start signal
	// return
	m.Lock()
	defer m.Unlock()

	if _, ok := m.services[name]; !ok {
		return errors.Wrapf(ErrBadRequest, "there is no tracked service with name %s", name)
	}

	service := m.services[name]
	if service.Status == Started || service.Status == Running || service.Status == Successful {
		return errors.Wrapf(ErrBadRequest, "service %s status is %s", name, service.Status)
	}
	m.startIfEligible(name)
	SminitLog.Info().Msgf("service %s is started", name)
	return nil
}

// Stop stops a service that is already tracked by the manager.
func (m *Manager) Stop(name string) error {
	// cancel service context.
	// return
	m.Lock()
	defer m.Unlock()

	if _, ok := m.services[name]; !ok {
		return errors.Wrapf(ErrBadRequest, "there is no tracked service with name %s", name)
	}

	service := m.services[name]
	service.stopSignal <- true
	return nil
}

// List lists all services tracked by the manager.
func (m *Manager) List() []Service {
	// list all services with their statuses
	// return
	m.Lock()
	defer m.Unlock()

	ret := make([]Service, 0)
	for _, service := range m.services {
		ret = append(ret, *service)
	}
	return ret
}

func (m *Manager) serviceRoutine(name string) {
	// watch for two signals: start and delete
	// start:
	//		continously start command
	// 		if healthcheck is provided, continously run health check
	//		if healthcheck is successful, change service status to Running
	//		if command terminated successfully, change service status to Successful
	//		a status change should always notify dependent services if they were waiting on this service
	//		leave this loop with a permenant error only if context is cancelled.
	// delete:
	//		return from this routine
	// default:
	//		do nothing
	service := m.services[name]
	for {
		select {
		case <-service.deleteSignal:
			return

		case <-service.startSignal:
			ctx, cancel := context.WithCancel(context.Background())
			go cancellationRoutine(cancel, service.stopSignal)

			err := backoff.Retry(func() error {
				select {

				case <-ctx.Done():
					service.isHealthy = false
					return backoff.Permanent(fmt.Errorf("service %s was stopped", service.Name))

				default:
					splittedCmd := strings.Split(service.cmdStr, " ")
					cmd := exec.CommandContext(ctx, splittedCmd[0], splittedCmd[1:]...)
					if service.log == "stdout" {
						cmd.Stdout = service.stdoutLogger
						cmd.Stderr = service.stderrLogger
					}

					err := cmd.Start()
					if err != nil {
						SminitLog.Error().Msgf("error while starting process %s. %s", service.Name, err.Error())
						return errors.New("restarting service")
					}
					service.Status = Started

					if isHealthy(ctx, service) {
						service.isHealthy = true
						service.Status = Running
						m.startEligibleChildren(service.Name)
					} else {
						service.isHealthy = false
					}

					err = cmd.Wait()
					if err != nil {
						SminitLog.Error().Msgf("error while running process %s. %s", service.Name, err.Error())
						service.isHealthy = false
						service.Status = Failure
						return errors.New("restarting service")
					} else {
						service.Status = Successful
						if service.oneShot {
							return backoff.Permanent(fmt.Errorf("service %s has finished", service.Name))
						}
					}

					return errors.New("restarting service")
				}
			}, newExponentialBackOff())

			SminitLog.Info().Msg(err.Error())

			if service.oneShot && service.Status == Successful {
				return
			}
			service.Status = Stopped

		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func cancellationRoutine(cancel context.CancelFunc, stopSignal chan bool) {
	<-stopSignal
	cancel()
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

func (m *Manager) startEligibleChildren(serviceName string) {
	// check if dependent services are eligible to be run
	for child := range m.services[serviceName].children {
		m.startIfEligible(child)
	}
}

func (m *Manager) startIfEligible(serviceName string) {
	service, ok := m.services[serviceName]

	if !ok || service.hasStarted() {
		return
	}

	startSignal := true
	for parentName := range service.parents {
		// if a parent is not present or not healthy, we cannot start service
		parent, ok := m.services[parentName]
		if !ok || !parent.isHealthy {
			startSignal = false
			break
		}
	}
	if startSignal {
		service.startSignal <- true
	}
}

func (s *Service) hasStarted() bool {
	return s.Status == Running || s.Status == Started || s.Status == Successful || s.Status == Failure
}

func generateService(service ServiceOptions) *Service {
	stdout := log.Output(zerolog.ConsoleWriter{
		Out: os.Stdout,
		FieldsExclude: []string{
			"component",
		},
		PartsOrder: []string{
			"level",
			"component",
			"message",
		},
		FormatLevel: func(i interface{}) string {
			return "INF"
		},
	}).With().Str("component", fmt.Sprintf("%s:", service.Name)).Logger().Level(zerolog.InfoLevel)

	stderr := log.Output(zerolog.ConsoleWriter{
		Out: os.Stdout,
		FieldsExclude: []string{
			"component",
		},
		PartsOrder: []string{
			"level",
			"component",
			"message",
		},
		FormatLevel: func(i interface{}) string {
			return "ERR"
		},
	}).With().Str("component", fmt.Sprintf("%s:", service.Name)).Logger().Level(zerolog.ErrorLevel)

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
		stdoutLogger: stdout,
		stderrLogger: stderr,
		children:     map[string]bool{},
		parents:      map[string]bool{},
		startSignal:  make(chan bool),
		stopSignal:   make(chan bool),
		deleteSignal: make(chan bool),
		isHealthy:    false,
	}
	return &newService
}

func (m *Manager) addToGraph(service ServiceOptions) error {
	for _, parent := range service.After {
		if _, ok := m.services[parent]; !ok {
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
