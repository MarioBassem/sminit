package swatcher

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
)

var wg sync.WaitGroup

// start's job is to traverse services and start them in order
func (s Swatcher) Start() error {

	for _, service := range s.SortedServices {
		wg.Add(1)
		go SpawnAndWatch(service)

	}
	wg.Wait()
	return nil
}

func SpawnAndWatch(service Service) error {
	err := backoff.Retry(func() error {
		log.Printf("[+]sminit: starting service %s", service.Name)
		s := strings.Split(service.CmdStr, " ")
		if len(s) < 1 {
			return backoff.Permanent(errors.New("cmd string is empty"))
		}
		cmd := exec.Command(s[0], s[:]...)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout

		err := cmd.Run()
		if err != nil {
			log.Printf("[-]sminit: error running process %s. %s", service.Name, err.Error())
			return errors.New("restarting service")
		}
		log.Printf("[+]sminit: service %s has finished. restarting...", service.Name)
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
		MaxInterval:         5 * time.Second,
		MaxElapsedTime:      0,
		Clock:               backoff.SystemClock,
	}
	b.Reset()
	return &b
}
