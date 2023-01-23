package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/mariobassem/sminit-go/internal/manager"
)

func StartHandler(args []string) {
	response, err := http.Post(fmt.Sprintf("http://%s:%d/start", manager.Address, manager.Port), "text", bytes.NewBuffer([]byte(args[0])))
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
