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
	Pending    Status = "pending"
	Started    Status = "started"
	Stopped    Status = "stopped"
)

type Service struct {
	name         string
	startSignal  chan bool
	deleteSignal chan bool
	status       Status
	// children are services that depend on this service.
	children map[string]bool
	// parents are services that this service depend on.
	parents     map[string]bool
	log         string
	healthCheck string
	oneShot     bool
	cmdStr      string
	context     context.Context
	cancel      context.CancelFunc
	stdout      *Stdout
	stderr      *Stderr
}

func NewManager(serviceOptions map[string]ServiceOptions) (Manager, error) {
	// generate map[string]Service
	// fire go routine for each service
	// services that have no parents should receive a start signal
	// return

	services := map[string]*Service{}
	manager := Manager{
		services: services,
	}
	for name, service := range serviceOptions {
		newService := generateService(service)
		manager.services[name] = newService
	}

	for _, service := range serviceOptions {
		err := manager.addToGraph(service)
		if err != nil {
			return Manager{}, errors.Wrap(err, "failed to modify service graph")
		}
	}

	for name, service := range manager.services {
		go manager.serviceRoutine(name)
		if len(service.parents) == 0 {
			service.startSignal <- true
		}
	}

	return manager, nil
}

// Add adds a new service to the list of services tracked by the manager
func (m *Manager) Add(opt ServiceOptions) error {
	// generate Service struct
	// fire go routine for service
	// check if all parents are in Running or Successful state, if true, send a start signal for this service
	// return
	if _, ok := m.services[opt.Name]; ok {
		return fmt.Errorf("failed to add %s. a service with the same name is already tracked.", opt.Name)
	}
	newService := generateService(opt)
	m.services[newService.name] = newService
	err := m.addToGraph(opt)
	if err != nil {
		return errors.Wrap(err, "failed to modify service graph")
	}

	go m.serviceRoutine(newService.name)
	m.startIfEligible(newService.name)

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
	service.cancel()
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
	if service.status == Started || service.status == Running || service.status == Successful {
		return fmt.Errorf("service %s status is %s", name, service.status)
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
	service.cancel()
	return nil
}

// List lists all services tracked by the manager.
func (m *Manager) List() []ServiceShort {
	// list all services with their statuses
	// return
	ret := []ServiceShort{}
	for name, service := range m.services {
		ret = append(ret, ServiceShort{
			Name:   name,
			Status: service.status,
		})
	}
	return ret
}

type ServiceShort struct {
	Name   string
	Status Status
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

			err := backoff.Retry(func() error {

				select {
				case <-service.context.Done():
					return backoff.Permanent(fmt.Errorf("service %s was stopped", service.name))
				default:

					cmd := exec.CommandContext(service.context, "bash", "-c", service.cmdStr)
					if service.log == "stdout" {
						cmd.Stdout = service.stdout
						cmd.Stderr = service.stderr
					}

					err := cmd.Start()
					if err != nil {
						SminitLogFail.Printf("error while starting process %s. %s", service.name, err.Error())
						return errors.New("restarting service")
					}
					m.changeStatus(service.name, Started)

					if isHealthy(service) {
						m.changeStatus(service.name, Running)
					}

					err = cmd.Wait()
					if err != nil {
						SminitLogFail.Printf("error while running process %s. %s", service.name, err.Error())
						return errors.New("restarting service")
					} else {
						m.changeStatus(service.name, Successful)
						if service.oneShot {
							return backoff.Permanent(fmt.Errorf("service %s has finished.", service.name))
						}
					}

					return errors.New("restarting service")
				}
			}, NewExponentialBackOff())

			SminitLog.Print(err)

			if service.oneShot && service.status == Successful {
				return
			}

			m.changeStatus(service.name, Stopped)

			context, cancel := context.WithCancel(context.Background())
			service.context = context
			service.cancel = cancel

		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// this function will return false only if context was cancelled, and true if cmd.Run() returned nil, i.e process is healthy
func isHealthy(service *Service) bool {
	for {
		select {
		case <-service.context.Done():
			return false
		default:
			cmd := exec.CommandContext(service.context, "bash", "-c", service.healthCheck)
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
	m.services[serviceName].status = newStatus
	for child := range m.services[serviceName].children {
		m.startIfEligible(child)
	}
}

func (m *Manager) startIfEligible(serviceName string) {
	service := m.services[serviceName]

	if service.status == Running || service.status == Started || service.status == Successful {
		return
	}
	startSignal := true
	for parent := range service.parents {
		if m.services[parent].status == Pending || m.services[parent].status == Stopped {
			startSignal = false
			break
		}
	}
	if startSignal {
		service.startSignal <- true
	}
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
	context, cancel := context.WithCancel(context.Background())
	healthCheck := service.HealthCheck
	if healthCheck == "" {
		healthCheck = "sleep 1"
	}

	newService := Service{
		name:         service.Name,
		status:       Pending,
		log:          service.Log,
		healthCheck:  healthCheck,
		cmdStr:       service.Cmd,
		oneShot:      service.OneShot,
		context:      context,
		cancel:       cancel,
		stdout:       &stdout,
		stderr:       &stderr,
		children:     map[string]bool{},
		parents:      map[string]bool{},
		startSignal:  make(chan bool),
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
