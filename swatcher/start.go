package swatcher

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
)

var wg sync.WaitGroup

// start's job is to traverse services and start them in order
func (s Swatcher) Start() error {

	for _, service := range s.SortedServices {
		if service.Log == "stdout" {
			service.Stdout = Stdout{
				File:   os.Stdout,
				Prefix: fmt.Sprintf("[+]%s: ", service.Name),
			}
			service.Stderr = Stderr{
				File:   os.Stderr,
				Prefix: fmt.Sprintf("[-]%s: ", service.Name),
			}
		}
		wg.Add(1)
		go SpawnAndWatch(service)

	}
	wg.Wait()
	return nil
}

func SpawnAndWatch(service Service) error {
	sminitLog := log.New(os.Stdout, "[+]sminit:", 0)
	sminitLogFail := log.New(os.Stdout, "[-]sminit:", 0)
	err := backoff.Retry(func() error {
		sminitLog.Printf("starting service %s", service.Name)

		cmd := createCommand(service)

		err := cmd.Run()
		if err != nil {
			sminitLogFail.Printf("error running process %s. %s", service.Name, err.Error())
			return errors.New("restarting service")
		}

		sminitLog.Printf("service %s has finished. restarting...", service.Name)
		return errors.New("restarting service")

	}, NewExponentialBackOff())

	wg.Done()
	if err != nil {
		return fmt.Errorf("[-]sminit: error running process %s. %s", service.Name, err.Error())
	}

	return nil
}

func NewExponentialBackOff() *backoff.ExponentialBackOff {
	b := backoff.ExponentialBackOff{
		InitialInterval:     backoff.DefaultInitialInterval,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
		Multiplier:          backoff.DefaultMultiplier,
		MaxInterval:         3 * time.Second,
		MaxElapsedTime:      0,
		Clock:               backoff.SystemClock,
	}
	b.Reset()
	return &b
}

func createCommand(service Service) *exec.Cmd {
	cmd := exec.Command("bash", "-c", service.Cmd)
	if service.Log == "stdout" {
		cmd.Stdout = &service.Stdout
		cmd.Stderr = &service.Stderr
	}
	return cmd
}
