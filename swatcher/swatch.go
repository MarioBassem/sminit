package swatcher

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/mariobassem/sminit-go/loader"
	"github.com/mariobassem/sminit-go/manager"
	"github.com/pkg/errors"
)

type Swatcher struct {
	Manager  manager.Manager
	Listener net.Listener
}

const (
	SwatchPidPath        = "/run/sminit/swatch.pid"
	SminitRunDir         = "/run/sminit"
	SwatchSocketPath     = "/run/sminit/swatch.sock"
	ServiceDefinitionDir = "/etc/sminit"
)

var (
	SminitLog     = log.New(os.Stdout, "[+]sminit:", 0)
	SminitLogFail = log.New(os.Stdout, "[-]sminit:", 0)
)

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

	listener, err := createSwatchSocket()
	if err != nil {
		return Swatcher{}, errors.Wrap(err, "couldn't create swatch socket")
	}

	services, err := loader.LoadAll(ServiceDefinitionDir)
	if err != nil {
		return Swatcher{}, err
		// do something
	}

	manager, err := manager.NewManager(services)
	if err != nil {
		return Swatcher{}, err
		// do sth
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

func createSwatchSocket() (net.Listener, error) {
	listener, err := net.Listen("unix", SwatchSocketPath)
	if err != nil {
		return nil, err
	}
	return listener, nil
}

func Swatch() error {
	swatcher, err := NewSwatcher()
	if err != nil {
		return err
		// do something
	}
	swatcher.StartServer()

	return nil

}

type Message struct {
	Success bool
	Content []byte
}

func (s *Swatcher) StartServer() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			SminitLogFail.Fatalf("failed to accept connection. %s", err.Error())
		}
		go func(conn net.Conn) {
			defer conn.Close()
			// Create a buffer for incoming data.
			buf := make([]byte, 1024)
			// Read data from the connection.
			n, err := conn.Read(buf)
			if err != nil {
				log.Fatal(err)
			}
			str := strings.Trim(string(buf[:n]), " \n")
			splitStr := strings.Split(str, " ")

			message := s.execute(splitStr)

			b, errMarshal := json.Marshal(message)
			if errMarshal != nil {
				SminitLogFail.Print(errMarshal)
			}

			_, errWrite := conn.Write(b)
			if errWrite != nil {
				SminitLogFail.Print(errWrite)
			}
			return

		}(conn)
	}
}

func (s *Swatcher) execute(splitStr []string) Message {
	cmd := splitStr[0]
	var message Message
	if cmd == "add" {
		message = s.handleAdd(splitStr[1:])
	} else if cmd == "delete" {
		message = s.handleDelete(splitStr[1:])
	} else if cmd == "start" {
		message = s.handleStart(splitStr[1:])
	} else if cmd == "stop" {
		message = s.handleStop(splitStr[1:])
	} else if cmd == "list" {
		message = s.handleList(splitStr[1:])
	} else {
		message = Message{
			Success: false,
			Content: []byte("wrong parametrs"),
		}
	}

	return message
}

func (s *Swatcher) handleAdd(args []string) Message {
	serviceName := args[0]
	service, err := loader.Load(ServiceDefinitionDir, serviceName)
	if err != nil {
		return Message{
			Success: false,
			Content: []byte(err.Error()),
		}
	}
	err = s.Manager.Add(service)
	if err != nil {
		return Message{
			Success: false,
			Content: []byte(err.Error()),
		}
	}
	return Message{
		Success: true,
	}
}

func (s *Swatcher) handleDelete(args []string) Message {
	serviceName := args[0]
	err := s.Manager.Delete(serviceName)
	if err != nil {
		return Message{
			Success: false,
			Content: []byte(err.Error()),
		}
	}
	return Message{
		Success: true,
	}
}

func (s *Swatcher) handleStart(args []string) Message {
	serviceName := args[0]
	err := s.Manager.Start(serviceName)
	if err != nil {
		return Message{
			Success: false,
			Content: []byte(err.Error()),
		}
	}
	return Message{
		Success: true,
	}
}

func (s *Swatcher) handleStop(args []string) Message {
	serviceName := args[0]
	err := s.Manager.Stop(serviceName)
	if err != nil {
		return Message{
			Success: false,
			Content: []byte(err.Error()),
		}
	}
	return Message{
		Success: true,
	}
}

func (s *Swatcher) handleList(args []string) Message {
	if len(args) != 0 {
		return Message{
			Success: false,
			Content: []byte("wrong parameters"),
		}
	}
	services := s.Manager.List()
	b, err := json.Marshal(services)
	if err != nil {
		return Message{
			Success: false,
			Content: []byte(err.Error()),
		}
	}
	return Message{
		Success: true,
		Content: b,
	}
}
