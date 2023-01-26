package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/mariobassem/sminit-go/internal/manager"
)

func ListHandler() {
	response, err := http.Get(fmt.Sprintf("http://%s:%d/services", manager.Address, manager.Port))
	if err != nil {
		manager.SminitLog.Error().Msgf("error sending list request: %s", err.Error())
		return
	}

	if response.StatusCode != http.StatusOK {
		manager.SminitLog.Error().Msgf("failure: %s", response.Status)
		return
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		manager.SminitLog.Error().Msgf("error reading sminit response body: %s", err.Error())
		return
	}
	services := []manager.Service{}
	err = json.Unmarshal(body, &services)
	if err != nil {
		manager.SminitLog.Error().Msgf("failed to unmarshal message content. %s", err.Error())
		return
	}
	manager.SminitLog.Info().Msg("tracked services:")
	for idx := range services {
		// TODO: needs to be changed
		log.Default().Writer().Write([]byte(fmt.Sprintf("\tname: %s, status: %s\n", services[idx].Name, services[idx].Status)))
		// log.Printf("\tname: %s, status: %s\n", services[idx].Name, services[idx].Status)
	}
}
