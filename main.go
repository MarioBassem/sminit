package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/mariobassem/sminit-go/manager"
	"github.com/mariobassem/sminit-go/swatcher"
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
			cntxt := &daemon.Context{
				// PidFileName: "sample2.pid",
				// PidFilePerm: 0644,
				// LogFileName: "sample2.log",
				LogFilePerm: 0640,
				WorkDir:     "/",
				Umask:       027,
				// Args:    []string{"[go-daemon sample2]"},
				LogFileName: "/run/sminit.log",
			}

			d, err := cntxt.Reborn()
			if err != nil {
				log.Fatal("Unable to run: ", err)
			}
			if d != nil {
				return
			}
			defer cntxt.Release()

			err = swatcher.Swatch()
			if err != nil {
				swatcher.SminitLogFail.Printf("error while executing Swatch. %s", err.Error())
			}
			defer swatcher.CleanUp()
		},
		Short: "Start a process that starts and watches all services defined in /etc/sminit",
		Args:  cobra.ExactArgs(0),
	}

	var startCmd = &cobra.Command{
		Use: "start",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := swatcher.NewClient()
			if err != nil {
				panic(err)
			}
			writeStr := strings.Join([]string{"start", args[0]}, " ")
			err = client.Write([]byte(writeStr))
			if err != nil {
				panic(err)
			}
			message, err := client.Read()
			if err != nil {
				panic(err)
			}
			// handle message

			if !message.Success {
				log.Fatalf("failure: %s", string(message.Content))
			}
		},
		Short: "Start a service that is already watched by sminit",
		Args:  cobra.ExactArgs(1),
	}

	var listCmd = &cobra.Command{
		Use: "list",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := swatcher.NewClient()
			if err != nil {
				panic(err)
			}
			writeStr := strings.Join([]string{"list"}, " ")
			err = client.Write([]byte(writeStr))
			if err != nil {
				panic(err)
			}
			message, err := client.Read()
			if err != nil {
				panic(err)
			}

			if !message.Success {
				log.Fatalf("failure: %s", string(message.Content))
			} else {
				services := make([]manager.ServiceShort, 10)
				err := json.Unmarshal(message.Content, &services)
				if err != nil {
					panic(err)
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
			client, err := swatcher.NewClient()
			if err != nil {
				panic(err)
			}
			writeStr := strings.Join([]string{"add", args[0]}, " ")
			err = client.Write([]byte(writeStr))
			if err != nil {
				panic(err)
			}
			message, err := client.Read()
			if err != nil {
				panic(err)
			}
			if !message.Success {
				log.Fatalf("failure: %s", string(message.Content))
			}
		},
		Short: "Add a new service that has a definition file in /etc/sminit to the services watched by sminit",
		Args:  cobra.ExactArgs(1),
	}

	var deleteCmd = &cobra.Command{
		Use: "delete [service_name]",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := swatcher.NewClient()
			if err != nil {
				panic(err)
			}
			writeStr := strings.Join([]string{"delete", args[0]}, " ")
			err = client.Write([]byte(writeStr))
			if err != nil {
				panic(err)
			}
			message, err := client.Read()
			if err != nil {
				panic(err)
			}
			if !message.Success {
				log.Fatalf("failure: %s", string(message.Content))
			}
		},
		Short: "Drop a service from the list of services that are being watched by sminit",
		Args:  cobra.ExactArgs(1),
	}

	var stopCmd = &cobra.Command{
		Use: "stop [service_name]",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := swatcher.NewClient()
			if err != nil {
				panic(err)
			}
			writeStr := strings.Join([]string{"stop", args[0]}, " ")
			err = client.Write([]byte(writeStr))
			if err != nil {
				panic(err)
			}
			message, err := client.Read()
			if err != nil {
				panic(err)
			}
			if !message.Success {
				log.Fatalf("failure: %s", string(message.Content))
			}
		},
		Short: "Stop a running service",
		Args:  cobra.ExactArgs(1),
	}

	var logCmd = &cobra.Command{
		Use: "log",
		Run: func(cmd *cobra.Command, args []string) {

			t, err := tail.TailFile(swatcher.SminitLogPath, tail.Config{Follow: true})
			if err != nil {
				panic(err)
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
	rootCmd.Execute()
}
