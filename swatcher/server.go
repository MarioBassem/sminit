package swatcher

import (
	"encoding/json"
	"net"
	"path"
	"strings"

	"github.com/mariobassem/sminit-go/loader"
)

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
			buf := make([]byte, 1024)

			n, err := conn.Read(buf)
			if err != nil {
				SminitLogFail.Print(err)
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
		message = s.handleList()
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
	fullName := strings.Join([]string{serviceName, ".yaml"}, "")
	path := path.Join(ServiceDefinitionDir, fullName)
	service, err := loader.Load(path, serviceName)
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

func (s *Swatcher) handleList() Message {
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
