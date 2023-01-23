package manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var (
	Address = "127.0.0.1"
	Port    = 8080
)

func (s *Watcher) startHTTPServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/start", s.start)
	mux.HandleFunc("/stop", s.stop)
	mux.HandleFunc("/add", s.add)
	mux.HandleFunc("/delete", s.delete)
	mux.HandleFunc("/list", s.list)
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", Address, Port), mux)
	return err
}

func (s *Watcher) start(w http.ResponseWriter, r *http.Request) {
	handler(w, r, s.Manager.Start)
}

func (s *Watcher) stop(w http.ResponseWriter, r *http.Request) {
	handler(w, r, s.Manager.Stop)
}

func (s *Watcher) delete(w http.ResponseWriter, r *http.Request) {
	handler(w, r, s.Manager.Delete)
}

func (s *Watcher) add(w http.ResponseWriter, r *http.Request) {
	handler(w, r, s.Manager.Add)
}

func (s *Watcher) list(w http.ResponseWriter, r *http.Request) {

	services := s.Manager.List()
	contentBytes, err := json.Marshal(services)
	if err != nil {
		SminitLog.Error().Msgf("could not marshal services: %+v. %s", services, err.Error())
		return
	}

	_, err = w.Write(contentBytes)
	if err != nil {
		SminitLog.Error().Msg(err.Error())
	}
}

func handler(w http.ResponseWriter, r *http.Request, action func(serviceName string) error) {
	serviceName, err := io.ReadAll(r.Body)
	if err != nil {
		SminitLog.Error().Msgf("could not read body: %s\n", err)
		return
	}
	defer r.Body.Close()

	err = action(string(serviceName))

	switch {
	case errors.Is(err, ErrBadRequest):
		w.WriteHeader(400)

	case errors.Is(err, ErrSminitInternalError):
		w.WriteHeader(500)

	}

	if err != nil {
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			SminitLog.Error().Msg(err.Error())
		}
	}

}
