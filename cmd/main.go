package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	swatch "github.com/mariobassem/sminit-go/pkg/swatcher"
	"github.com/nxadm/tail"
	"github.com/sevlyar/go-daemon"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:       "sminit [subcommand]",
		Short:     "sminit is a trivial service manager",
		Example:   "sminit start service_name",
		ValidArgs: []string{"swatch", "start", "stop", "add", "delete", "list"},
	}

	var swatchCmd = &cobra.Command{
		Use: "swatch",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := &daemon.Context{
				// PidFileName: "sample2.pid",
				// PidFilePerm: 0644,
				// LogFileName: "sample2.log",
				LogFilePerm: 0640,
				WorkDir:     "/",
				Umask:       027,
				// Args:    []string{"[go-daemon sample2]"},
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

			err = swatch.Swatch()
			if err != nil {
				swatch.SminitLog.Error().Msgf("error while executing Swatch. %s", err.Error())
			}
			defer swatch.CleanUp()
		},
		Short: "Start a process that starts and watches all services defined in /etc/sminit",
		Args:  cobra.ExactArgs(0),
	}

	var startCmd = &cobra.Command{
		Use: "start",
		Run: func(cmd *cobra.Command, args []string) {

			response, err := http.Post(fmt.Sprintf("http://%s:%d/start", swatch.Address, swatch.Port), "text", bytes.NewBuffer([]byte(args[0])))
			if err != nil {
				swatch.SminitLog.Error().Msgf("error sending start request: %s", err.Error())
				return

			}
			if response.StatusCode == http.StatusOK {
				return
			}

			body, err := io.ReadAll(response.Body)
			if err != nil {
				swatch.SminitLog.Error().Msgf("error reading sminit response body: %s", err.Error())
				return
			}
			defer response.Body.Close()

			swatch.SminitLog.Error().Msgf("%s: %s", response.Status, string(body))

		},
		Short: "Start a service that is already watched by sminit",
		Args:  cobra.ExactArgs(1),
	}

	var listCmd = &cobra.Command{
		Use: "list",
		Run: func(cmd *cobra.Command, args []string) {
			response, err := http.Get(fmt.Sprintf("http://%s:%d/list", swatch.Address, swatch.Port))
			if err != nil {
				swatch.SminitLog.Error().Msgf("error sending list request: %s", err.Error())
				return
			}

			if response.StatusCode != http.StatusOK {
				swatch.SminitLog.Error().Msgf("failure: %s", response.Status)
				return
			}

			body, err := io.ReadAll(response.Body)
			if err != nil {
				swatch.SminitLog.Error().Msgf("error reading sminit response body: %s", err.Error())
				return
			}
			services := []swatch.Service{}
			err = json.Unmarshal(body, &services)
			if err != nil {
				swatch.SminitLog.Error().Msgf("failed to unmarshal message content. %s", err.Error())
				return
			}
			swatch.SminitLog.Info().Msg("tracked services:")
			for idx := range services {
				log.Default().Writer().Write([]byte(fmt.Sprintf("\tname: %s, status: %s\n", services[idx].Name, services[idx].Status)))
				// log.Printf("\tname: %s, status: %s\n", services[idx].Name, services[idx].Status)
			}
		},
		Short: "List all services that are watched by sminit",
		Args:  cobra.ExactArgs(0),
	}

	var addCmd = &cobra.Command{
		Use: "add [service_name]",
		Run: func(cmd *cobra.Command, args []string) {
			response, err := http.Post(fmt.Sprintf("http://%s:%d/add", swatch.Address, swatch.Port), "text", bytes.NewBuffer([]byte(args[0])))
			if err != nil {
				swatch.SminitLog.Error().Msgf("error sending add request: %s", err.Error())
				return
			}
			if response.StatusCode == http.StatusOK {
				return
			}

			body, err := io.ReadAll(response.Body)
			if err != nil {
				swatch.SminitLog.Error().Msgf("error reading sminit response body: %s", err.Error())
				return
			}
			defer response.Body.Close()

			swatch.SminitLog.Error().Msgf("%s: %s", response.Status, string(body))
		},
		Short: "Add a new service that has a definition file in /etc/sminit to the services watched by sminit",
		Args:  cobra.ExactArgs(1),
	}

	var deleteCmd = &cobra.Command{
		Use: "delete [service_name]",
		Run: func(cmd *cobra.Command, args []string) {
			response, err := http.Post(fmt.Sprintf("http://%s:%d/delete", swatch.Address, swatch.Port), "text", bytes.NewBuffer([]byte(args[0])))
			if err != nil {
				swatch.SminitLog.Error().Msgf("error sending delete request: %s", err.Error())
				return
			}
			if response.StatusCode == http.StatusOK {
				return
			}

			body, err := io.ReadAll(response.Body)
			if err != nil {
				swatch.SminitLog.Error().Msgf("error reading sminit response body: %s", err.Error())
				return
			}
			defer response.Body.Close()

			swatch.SminitLog.Error().Msgf("%s: %s", response.Status, string(body))
		},
		Short: "Drop a service from the list of services that are being watched by sminit",
		Args:  cobra.ExactArgs(1),
	}

	var stopCmd = &cobra.Command{
		Use: "stop [service_name]",
		Run: func(cmd *cobra.Command, args []string) {
			response, err := http.Post(fmt.Sprintf("http://%s:%d/stop", swatch.Address, swatch.Port), "text", bytes.NewBuffer([]byte(args[0])))
			if err != nil {
				swatch.SminitLog.Error().Msgf("error sending stop request: %s", err.Error())
				return
			}
			if response.StatusCode == http.StatusOK {
				return
			}

			body, err := io.ReadAll(response.Body)
			if err != nil {
				swatch.SminitLog.Error().Msgf("error reading sminit response body: %s", err.Error())
				return
			}
			defer response.Body.Close()

			swatch.SminitLog.Error().Msgf("%s: %s", response.Status, string(body))
		},
		Short: "Stop a running service",
		Args:  cobra.ExactArgs(1),
	}

	var logCmd = &cobra.Command{
		Use: "log",
		Run: func(cmd *cobra.Command, args []string) {

			t, err := tail.TailFile(swatch.SminitLogPath, tail.Config{Follow: true})
			if err != nil {
				swatch.SminitLog.Error().Msgf("failed to print logs. %s", err.Error())
				return
			}
			for line := range t.Lines {
				fmt.Println(line.Text)
			}
		},
		Short: "Show sminit logs",
		Args:  cobra.ExactArgs(0),
	}

	rootCmd.AddCommand(swatchCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(logCmd)
	_ = rootCmd.Execute()
}
