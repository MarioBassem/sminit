package swatch

import (
	"fmt"
	"io/fs"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/pkg/errors"
)

type Swatcher struct {
	Manager  Manager
	Listener net.Listener
}

const (
	SwatchPidPath        = "/run/sminit/swatch.pid"
	SminitRunDir         = "/run/sminit"
	SminitLogPath        = "/run/sminit.log"
	SwatchSocketPath     = "/run/sminit/swatch.sock"
	ServiceDefinitionDir = "/etc/sminit"
)

var (
	SminitLog     = log.New(os.Stdout, "[+]sminit: ", log.Lmsgprefix)
	SminitLogFail = log.New(os.Stdout, "[-]sminit: ", log.Lmsgprefix)
)

func Swatch() error {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sigs
		CleanUp()
		os.Exit(0)
	}()

	swatcher, err := NewSwatcher()
	if err != nil {
		return errors.Wrap(err, "failed to create a new Swatcher")
	}
	swatcher.Start()

	return nil

}

// NewSwatcher creates a new Swatcher, it should return an error if an instance of Swatcher is already running.
func NewSwatcher() (Swatcher, error) {
	pid, err := getRunningInstance()
	if err == nil {
		return Swatcher{}, errors.New(fmt.Sprintf("there is a running instance of swatch with pid %d", pid))
	} else if !errors.Is(err, fs.ErrNotExist) {
		return Swatcher{}, errors.Wrap(err, "unexpected error")
	}

	err = os.Mkdir(SminitRunDir, fs.ModeDir)
	if err != nil {
		return Swatcher{}, errors.Wrapf(err, "couldn't create directory %s", SminitRunDir)
	}

	err = createSwatchPidFile()
	if err != nil {
		return Swatcher{}, errors.Wrap(err, "couldn't create swatch pid file")
	}

	listener, err := listen()
	if err != nil {
		return Swatcher{}, errors.Wrap(err, "couldn't create swatch socket")
	}

	services, err := LoadAll(ServiceDefinitionDir)
	if err != nil {
		return Swatcher{}, err
	}

	manager, err := NewManager(services)
	if err != nil {
		return Swatcher{}, err
	}

	return Swatcher{
		Listener: listener,
		Manager:  manager,
	}, nil
}

// getRunningInstance returns the pid of the running instance of swatch.
func getRunningInstance() (pid int, err error) {
	b, err := os.ReadFile(SwatchPidPath)
	if err != nil {
		return 0, errors.Wrap(err, "couldn't get swatch running instance pid")
	}

	pid, err = strconv.Atoi(string(b))
	if err != nil {
		return 0, errors.Wrapf(err, "couldn't convert bytes from %s to int", SwatchPidPath)
	}

	return pid, nil
}

func createSwatchPidFile() error {
	f, err := os.Create(SwatchPidPath)
	if err != nil {
		return errors.Wrapf(err, "couldn't create %s", SwatchPidPath)
	}

	pidBytes := []byte(strconv.FormatInt(int64(os.Getpid()), 10))
	_, err = f.Write(pidBytes)
	if err != nil {
		return errors.Wrapf(err, "couldn't write swatch pid in %s", SwatchPidPath)
	}

	return nil
}

func listen() (net.Listener, error) {
	listener, err := net.Listen("unix", SwatchSocketPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create a listener on socket %s", SwatchSocketPath)
	}
	return listener, nil
}

// CleanUp should delete /run/sminit directory and /run/sminit.log
func CleanUp() {
	err := os.RemoveAll(SminitRunDir)
	if err != nil {
		SminitLogFail.Printf("error while removing %s, you will have to remove it manually. %s", SminitRunDir, err.Error())
	}
	err = os.Remove(SminitLogPath)
	if err != nil {
		SminitLogFail.Printf("error while removing %s, you will have to remove it manually. %s", SminitLogPath, err.Error())
	}
}
