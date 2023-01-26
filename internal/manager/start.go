package manager

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// App handles services through the manager, and accepts requests through Listener
type App struct {
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

// StartApp starts a daemon process responsible for tracking services, and exposing an http server that accepts requests to manipulate those services
func StartApp() error {
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

	watcher := App{
		Manager:  manager,
		Listener: listener,
	}

	err = watcher.startHTTPServer()
	if err != nil {
		SminitLog.Error().Msgf("error starting http server: %s", err)
	}

	return nil

}
