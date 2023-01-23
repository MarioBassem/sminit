package manager

import (
	"fmt"
	"io/fs"

	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Watcher handles services through the manager, and accepts requests through Listener
type Watcher struct {
	Manager  *Manager
	Listener net.Listener
}

const (
	SminitPidPath        = "/run/sminit/sminit.pid"
	SminitRunDir         = "/run/sminit"
	SminitLogPath        = "/run/sminit.log"
	SminitSocketPath     = "/run/sminit/sminit.sock"
	ServiceDefinitionDir = "/etc/sminit"
)

var (
	// SminitLog is the default logger used in sminit
	SminitLog = log.Output(zerolog.ConsoleWriter{
		Out: os.Stdout,
		FieldsExclude: []string{
			"component",
		},
		PartsOrder: []string{
			"level",
			"component",
			"message",
		},
	}).With().Str("component", "sminit:").Logger()
)

// Watch starts a daemon process responsible for tracking services, and exposing an http server that accepts requests to manipulate those services
func Watch() error {
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

	listener, err := net.Listen("unix", SminitSocketPath)
	if err != nil {
		return errors.Wrapf(err, "failed to create a listener on socket %s", SminitSocketPath)
	}

	services, err := LoadAll(ServiceDefinitionDir)
	if err != nil {
		return err
	}

	manager, err := NewManager(services)
	if err != nil {
		return err
	}
	manager.fireServices()

	watcher := Watcher{
		Manager:  manager,
		Listener: listener,
	}

	err = watcher.startHTTPServer()
	if err != nil {
		SminitLog.Error().Msgf("error starting http server: %s", err)
	}

	return nil

}

func createFilesAndDirs() error {
	pid, err := getRunningInstance()
	if err == nil {
		return errors.New(fmt.Sprintf("there is a running instance of sminit with pid %d", pid))
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	err = os.Mkdir(SminitRunDir, fs.ModeDir)
	if err != nil {
		return errors.Wrapf(err, "could not create directory %s", SminitRunDir)
	}

	err = createSminitPidFile()
	if err != nil {
		return errors.Wrap(err, "could not create sminit pid file")
	}
	return nil
}

// getRunningInstance returns the pid of the running instance of sminit.
func getRunningInstance() (pid int, err error) {
	b, err := os.ReadFile(SminitPidPath)
	if err != nil {
		return 0, errors.Wrapf(err, "could not read file %s", SminitPidPath)
	}

	pid, err = strconv.Atoi(string(b))
	if err != nil {
		return 0, errors.Wrapf(err, "could not convert bytes from %s to int", SminitPidPath)
	}

	return pid, nil
}

func createSminitPidFile() error {
	f, err := os.Create(SminitPidPath)
	if err != nil {
		return errors.Wrapf(err, "could not create %s", SminitPidPath)
	}

	pidBytes := []byte(strconv.FormatInt(int64(os.Getpid()), 10))
	_, err = f.Write(pidBytes)
	if err != nil {
		return err
	}

	return nil
}

// CleanUp should delete /run/sminit directory and /run/sminit.log
func CleanUp() {
	err := os.RemoveAll(SminitRunDir)
	if err != nil {
		SminitLog.Error().Msgf("error while removing %s, you need to remove it manually. %s", SminitRunDir, err.Error())
	}
	err = os.Remove(SminitLogPath)
	if err != nil {
		SminitLog.Error().Msgf("error while removing %s, you need to remove it manually. %s", SminitLogPath, err.Error())
	}
}
