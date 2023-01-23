package handler

import (
	"fmt"

	"github.com/mariobassem/sminit-go/internal/manager"
	"github.com/nxadm/tail"
)

func LogHandler() {
	t, err := tail.TailFile(manager.SminitLogPath, tail.Config{Follow: true})
	if err != nil {
		manager.SminitLog.Error().Msgf("failed to print logs. %s", err.Error())
		return
	}
	for line := range t.Lines {
		fmt.Println(line.Text)
	}
}
