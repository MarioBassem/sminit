package handler

import (
	"log"

	"github.com/mariobassem/sminit-go/internal/manager"
	"github.com/sevlyar/go-daemon"
)

func InitHandler() {
	ctx := &daemon.Context{
		LogFilePerm: 0640,
		WorkDir:     "/",
		Umask:       027,
		LogFileName: "/run/sminit.log",
	}

	d, err := ctx.Reborn()
	if err != nil {
		log.Fatal("Unable to run: ", err)
	}
	if d != nil {
		return
	}
	defer ctx.Release()

	err = manager.StartApp()
	if err != nil {
		manager.SminitLog.Error().Msg(err.Error())
	}
	defer manager.CleanUp()
}
