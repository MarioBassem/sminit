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

	err := createFilesAndDirs()
	if err != nil {
		return errors.Wrap(err, "failed to create required files and directories")
	}

	listener, err := net.Listen("unix", SwatchSocketPath)
	if err != nil {
		return errors.Wrapf(err, "failed to create a listener on socket %s", SwatchSocketPath)
	}

	services, err := LoadAll(ServiceDefinitionDir)
	if err != nil {
		return err
	}

	manager, err := NewManager(services)
	if err != nil {
		return err
	}
	manager.FireServices()

	swatcher := Swatcher{
		Manager:  manager,
		Listener: listener,
	}

	swatcher.Start()

	return nil

}

func createFilesAndDirs() error {
	pid, err := getRunningInstance()
	if err == nil {
		return errors.New(fmt.Sprintf("there is a running instance of swatch with pid %d", pid))
	} else if !errors.Is(err, fs.ErrNotExist) {
		return errors.Wrap(err, "unexpected error")
	}

	err = os.Mkdir(SminitRunDir, fs.ModeDir)
	if err != nil {
		return errors.Wrapf(err, "could not create directory %s", SminitRunDir)
	}

	err = createSwatchPidFile()
	if err != nil {
		return errors.Wrap(err, "could not create swatch pid file")
	}
	return nil
}

// getRunningInstance returns the pid of the running instance of swatch.
func getRunningInstance() (pid int, err error) {
	b, err := os.ReadFile(SwatchPidPath)
	if err != nil {
		return 0, errors.Wrap(err, "could not get swatch running instance pid")
	}

	pid, err = strconv.Atoi(string(b))
	if err != nil {
		return 0, errors.Wrapf(err, "could not convert bytes from %s to int", SwatchPidPath)
	}

	return pid, nil
}

func createSwatchPidFile() error {
	f, err := os.Create(SwatchPidPath)
	if err != nil {
		return errors.Wrapf(err, "could not create %s", SwatchPidPath)
	}

	pidBytes := []byte(strconv.FormatInt(int64(os.Getpid()), 10))
	_, err = f.Write(pidBytes)
	if err != nil {
		return errors.Wrapf(err, "could not write swatch pid in %s", SwatchPidPath)
	}

	return nil
}

// CleanUp should delete /run/sminit directory and /run/sminit.log
func CleanUp() {
	err := os.RemoveAll(SminitRunDir)
	if err != nil {
		SminitLogFail.Printf("error while removing %s, you need to remove it manually. %s", SminitRunDir, err.Error())
	}
	err = os.Remove(SminitLogPath)
	if err != nil {
		SminitLogFail.Printf("error while removing %s, you need to remove it manually. %s", SminitLogPath, err.Error())
	}
}
