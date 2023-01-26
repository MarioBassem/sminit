package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/mariobassem/sminit-go/internal/manager"
)

func StartHandler(args []string) {
	client := &http.Client{}
	request, err := http.NewRequest(http.MethodPut, fmt.Sprintf("http://%s:%d/services/%s/start", manager.Address, manager.Port, args[0]), nil)
	if err != nil {
		manager.SminitLog.Error().Msgf("error creating start request: %s", err.Error())
		return
	}

	response, err := client.Do(request)
	if err != nil {
		manager.SminitLog.Error().Msgf("error sending start request: %s", err.Error())
		return

	}
	if response.StatusCode == http.StatusOK {
		return
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		manager.SminitLog.Error().Msgf("error reading sminit response body: %s", err.Error())
		return
	}
	defer response.Body.Close()

	manager.SminitLog.Error().Msgf("%s: %s", response.Status, string(body))

}
