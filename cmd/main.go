package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

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
				swatch.SminitLogFail.Printf("error while executing Swatch. %s", err.Error())
			}
			defer swatch.CleanUp()
		},
		Short: "Start a process that starts and watches all services defined in /etc/sminit",
		Args:  cobra.ExactArgs(0),
	}

	var startCmd = &cobra.Command{
		Use: "start",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := swatch.NewClient()
			if err != nil {
				swatch.SminitLogFail.Printf("error while creating a new client. %s", err.Error())
			}
			writeStr := strings.Join([]string{"start", args[0]}, " ")
			err = client.Write([]byte(writeStr))
			if err != nil {
				swatch.SminitLogFail.Printf("error while writing. %s", err.Error())
			}
			message, err := client.Read()
			if err != nil {
				swatch.SminitLogFail.Printf("error while reading returned message. %s", err.Error())
			}
			// handle message

			if !message.Success {
				swatch.SminitLogFail.Printf("failure: %s", string(message.Content))
			}
		},
		Short: "Start a service that is already watched by sminit",
		Args:  cobra.ExactArgs(1),
	}

	var listCmd = &cobra.Command{
		Use: "list",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := swatch.NewClient()
			if err != nil {
				swatch.SminitLogFail.Printf("error while creating a new client. %s", err.Error())
			}
			writeStr := strings.Join([]string{"list"}, " ")
			err = client.Write([]byte(writeStr))
			if err != nil {
				swatch.SminitLogFail.Printf("error while writing. %s", err.Error())
			}
			message, err := client.Read()
			if err != nil {
				swatch.SminitLogFail.Printf("error while reading returned message. %s", err.Error())
			}

			if !message.Success {
				log.Fatalf("failure: %s", string(message.Content))
			} else {
				services := make([]swatch.Service, 10)
				err := json.Unmarshal(message.Content, &services)
				if err != nil {
					swatch.SminitLogFail.Printf("failed to unmarshal message. %s", err.Error())
				}
				log.Print(services)
			}
		},
		Short: "List all services that are watched by sminit",
		Args:  cobra.ExactArgs(0),
	}

	var addCmd = &cobra.Command{
		Use: "add [service_name]",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := swatch.NewClient()
			if err != nil {
				swatch.SminitLogFail.Printf("error while creating a new client. %s", err.Error())
			}
			writeStr := strings.Join([]string{"add", args[0]}, " ")
			err = client.Write([]byte(writeStr))
			if err != nil {
				swatch.SminitLogFail.Printf("error while writing. %s", err.Error())
			}
			message, err := client.Read()
			if err != nil {
				swatch.SminitLogFail.Printf("error while reading returned message. %s", err.Error())
			}
			// handle message

			if !message.Success {
				swatch.SminitLogFail.Printf("failure: %s", string(message.Content))
			}
		},
		Short: "Add a new service that has a definition file in /etc/sminit to the services watched by sminit",
		Args:  cobra.ExactArgs(1),
	}

	var deleteCmd = &cobra.Command{
		Use: "delete [service_name]",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := swatch.NewClient()
			if err != nil {
				swatch.SminitLogFail.Printf("error while creating a new client. %s", err.Error())
			}
			writeStr := strings.Join([]string{"delete", args[0]}, " ")
			err = client.Write([]byte(writeStr))
			if err != nil {
				swatch.SminitLogFail.Printf("error while writing. %s", err.Error())
			}
			message, err := client.Read()
			if err != nil {
				swatch.SminitLogFail.Printf("error while reading returned message. %s", err.Error())
			}
			// handle message

			if !message.Success {
				swatch.SminitLogFail.Printf("failure: %s", string(message.Content))
			}
		},
		Short: "Drop a service from the list of services that are being watched by sminit",
		Args:  cobra.ExactArgs(1),
	}

	var stopCmd = &cobra.Command{
		Use: "stop [service_name]",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := swatch.NewClient()
			if err != nil {
				swatch.SminitLogFail.Printf("error while creating a new client. %s", err.Error())
			}
			writeStr := strings.Join([]string{"stop", args[0]}, " ")
			err = client.Write([]byte(writeStr))
			if err != nil {
				swatch.SminitLogFail.Printf("error while writing. %s", err.Error())
			}
			message, err := client.Read()
			if err != nil {
				swatch.SminitLogFail.Printf("error while reading returned message. %s", err.Error())
			}
			// handle message

			if !message.Success {
				swatch.SminitLogFail.Printf("failure: %s", string(message.Content))
			}
		},
		Short: "Stop a running service",
		Args:  cobra.ExactArgs(1),
	}

	var logCmd = &cobra.Command{
		Use: "log",
		Run: func(cmd *cobra.Command, args []string) {

			t, err := tail.TailFile(swatch.SminitLogPath, tail.Config{Follow: true})
			if err != nil {
				swatch.SminitLogFail.Printf("failed to print logs. %s", err.Error())
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
