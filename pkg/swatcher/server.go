package swatch

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

func (s *Swatcher) StartHTTPServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/start", s.startHandler)
	mux.HandleFunc("/stop", s.stopHandler)
	mux.HandleFunc("/add", s.addHandler)
	mux.HandleFunc("/delete", s.deleteHandler)
	mux.HandleFunc("/list", s.listHandler)
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", Address, Port), mux)
	return err
}

func (s *Swatcher) startHandler(w http.ResponseWriter, r *http.Request) {
	sminitHandler(w, r, s.Manager.Start)
}

func (s *Swatcher) stopHandler(w http.ResponseWriter, r *http.Request) {
	sminitHandler(w, r, s.Manager.Stop)
}

func (s *Swatcher) deleteHandler(w http.ResponseWriter, r *http.Request) {
	sminitHandler(w, r, s.Manager.Delete)
}

func (s *Swatcher) addHandler(w http.ResponseWriter, r *http.Request) {
	sminitHandler(w, r, s.Manager.Add)
}

func (s *Swatcher) listHandler(w http.ResponseWriter, r *http.Request) {

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

func sminitHandler(w http.ResponseWriter, r *http.Request, action func(serviceName string) error) {
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
		fallthrough
	case errors.Is(err, ErrSminitInternalError):
		w.WriteHeader(500)
		fallthrough
	case err != nil:
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			SminitLog.Error().Msg(err.Error())
		}
	}

}
