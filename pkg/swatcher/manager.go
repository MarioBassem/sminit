package swatch

import (
	"context"

	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"
)

// Manager handles service manipulation
type Manager struct {
	services map[string]*Service
}

type Status string

const (
	Running    Status = "running"
	Successful Status = "successful"
	Started    Status = "started"

	Pending Status = "pending"
	Failure Status = "failure"
	Stopped Status = "stopped"
)

type Service struct {
	Name   string
	Status Status

	startSignal  chan bool
	deleteSignal chan bool
	stopSignal   chan bool
	// children are services that depend on this service.
	children map[string]bool
	// parents are services that this service depend on.
	parents     map[string]bool
	log         string
	healthCheck string
	oneShot     bool
	cmdStr      string

	stdout *Stdout
	stderr *Stderr
}

// NewManager creates a new Manager struct and populates it with services generated from provided serviceOptions
func NewManager(serviceOptions map[string]ServiceOptions) (Manager, error) {
	manager := Manager{
		services: make(map[string]*Service),
	}

	err := manager.populateServices(serviceOptions)
	if err != nil {
		return Manager{}, errors.Wrap(err, "failed to populate manager with services")
	}

	return manager, nil
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

// FireServices is responsible for starting a go routine for each service, and starting it if eligible
func (m *Manager) FireServices() {
	for name := range m.services {
		go m.serviceRoutine(name)
		m.startIfEligible(name)
	}
}

// Add adds a new service to the list of services tracked by the manager
func (m *Manager) Add(opt ServiceOptions) error {
	// generate Service struct
	// fire go routine for service
	// check if all parents are in Running or Successful state, if true, send a start signal for this service
	// return
	if _, ok := m.services[opt.Name]; ok {
		return fmt.Errorf("failed to add %s. a service with the same name is already tracked", opt.Name)
	}
	newService := generateService(opt)
	m.services[newService.Name] = newService
	err := m.addToGraph(opt)
	if err != nil {
		return errors.Wrap(err, "failed to modify service graph")
	}

	go m.serviceRoutine(newService.Name)
	m.startIfEligible(newService.Name)

	return nil
}

// Delete deletes a services with the given name from the list of services tracked by the manager
func (m *Manager) Delete(name string) error {
	// cancel service context
	// send delete signal
	// remove from service map, and from graph
	// return
	if _, ok := m.services[name]; !ok {
		return fmt.Errorf("there is no tracked service with name %s", name)
	}

	service := m.services[name]
	service.stopSignal <- true
	service.deleteSignal <- true
	for parent := range service.parents {
		delete(m.services[parent].children, name)
	}
	for child := range service.children {
		delete(m.services[child].parents, name)
		m.startIfEligible(child)
	}
	delete(m.services, name)
	return nil
}

// Start starts a services that is already tracked by the manager.
func (m *Manager) Start(name string) error {
	// check parents' statuses of service
	// if all are running or successful, send start signal
	// return
	if _, ok := m.services[name]; !ok {
		return fmt.Errorf("there is no tracked service with name %s", name)
	}

	service := m.services[name]
	if service.Status == Started || service.Status == Running || service.Status == Successful {
		return fmt.Errorf("service %s status is %s", name, service.Status)
	}
	m.startIfEligible(name)
	return nil
}

// Stop stops a service that is already tracked by the manager.
func (m *Manager) Stop(name string) error {
	// cancel service context.
	// return
	if _, ok := m.services[name]; !ok {
		return fmt.Errorf("there is no tracked service with name %s", name)
	}

	service := m.services[name]
	service.stopSignal <- true
	return nil
}

// List lists all services tracked by the manager.
func (m *Manager) List() []Service {
	// list all services with their statuses
	// return
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
					return backoff.Permanent(fmt.Errorf("service %s was stopped", service.Name))

				default:

					cmd := exec.CommandContext(ctx, "bash", "-c", service.cmdStr)
					if service.log == "stdout" {
						cmd.Stdout = service.stdout
						cmd.Stderr = service.stderr
					}

					err := cmd.Start()
					if err != nil {
						SminitLogFail.Printf("error while starting process %s. %s", service.Name, err.Error())
						return errors.New("restarting service")
					}
					m.changeStatus(service.Name, Started)

					if isHealthy(ctx, service) {
						m.changeStatus(service.Name, Running)
					}

					err = cmd.Wait()
					if err != nil {
						SminitLogFail.Printf("error while running process %s. %s", service.Name, err.Error())
						m.changeStatus(service.Name, Failure)
						return errors.New("restarting service")
					} else {
						m.changeStatus(service.Name, Successful)
						if service.oneShot {
							return backoff.Permanent(fmt.Errorf("service %s has finished", service.Name))
						}
					}

					return errors.New("restarting service")
				}
			}, NewExponentialBackOff())

			SminitLog.Print(err)

			if service.oneShot && service.Status == Successful {
				return
			}

			m.changeStatus(service.Name, Stopped)

		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func cancellationRoutine(cancel context.CancelFunc, stopSignal chan bool) {
	for {
		select {
		case <-stopSignal:
			cancel()
			return
		default:
			continue
		}
	}
}

// this function will return false only if context was cancelled, and true if cmd.Run() returned nil, i.e process is healthy
func isHealthy(ctx context.Context, service *Service) bool {
	for {
		select {
		case <-ctx.Done():
			return false
		default:
			cmd := exec.CommandContext(ctx, "bash", "-c", service.healthCheck)
			err := cmd.Run()
			if err == nil {
				return true
			}
			time.Sleep(time.Millisecond * 500)
		}
	}
}

func (m *Manager) changeStatus(serviceName string, newStatus Status) {
	// change status of service
	// check if dependent services are eligible to be run
	m.services[serviceName].Status = newStatus
	for child := range m.services[serviceName].children {
		m.startIfEligible(child)
	}
}

func (m *Manager) startIfEligible(serviceName string) {
	service := m.services[serviceName]

	if service.hasStarted() {
		return
	}
	startSignal := true
	for parent := range service.parents {
		// if a parent is waiting, we cannot start service
		if m.services[parent].isNotHealthy() {
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

func (s *Service) isNotHealthy() bool {
	return s.Status == Pending || s.Status == Stopped
}

func generateService(service ServiceOptions) *Service {
	stdout := Stdout{
		File:   os.Stdout,
		Prefix: fmt.Sprintf("[+]%s: ", service.Name),
	}
	stderr := Stderr{
		File:   os.Stderr,
		Prefix: fmt.Sprintf("[-]%s: ", service.Name),
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
		stdout:       &stdout,
		stderr:       &stderr,
		children:     map[string]bool{},
		parents:      map[string]bool{},
		startSignal:  make(chan bool),
		stopSignal:   make(chan bool),
		deleteSignal: make(chan bool),
	}
	return &newService
}

func (m *Manager) addToGraph(service ServiceOptions) error {
	for _, parent := range service.After {
		if _, ok := m.services[parent]; !ok {
			return fmt.Errorf("service %s does not exist", parent)
		}
		m.services[service.Name].parents[parent] = true
		m.services[parent].children[service.Name] = true
	}
	return nil
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

func NewExponentialBackOff() *backoff.ExponentialBackOff {
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
