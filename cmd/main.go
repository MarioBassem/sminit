package main

import (
	handler "github.com/mariobassem/sminit-go/internal/handlers"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:       "sminit [subcommand]",
		Short:     "sminit is a trivial service manager",
		Example:   "sminit start service_name",
		ValidArgs: []string{"init", "start", "stop", "add", "delete", "list"},
	}

	var initCmd = &cobra.Command{
		Use: "init",
		Run: func(cmd *cobra.Command, args []string) {
			handler.InitHandler()
		},
		Short: "Start a process that starts and watches all services defined in /etc/sminit",
		Args:  cobra.ExactArgs(0),
	}

	var startCmd = &cobra.Command{
		Use: "start",
		Run: func(cmd *cobra.Command, args []string) {
			handler.StartHandler(args)
		},
		Short: "Start a service that is already watched by sminit",
		Args:  cobra.ExactArgs(1),
	}

	var listCmd = &cobra.Command{
		Use: "list",
		Run: func(cmd *cobra.Command, args []string) {
			handler.ListHandler()
		},
		Short: "List all services that are watched by sminit",
		Args:  cobra.ExactArgs(0),
	}

	var addCmd = &cobra.Command{
		Use: "add [service_name]",
		Run: func(cmd *cobra.Command, args []string) {
			handler.AddHandler(args)
		},
		Short: "Add a new service that has a definition file in /etc/sminit to the services watched by sminit",
		Args:  cobra.ExactArgs(1),
	}

	var deleteCmd = &cobra.Command{
		Use: "delete [service_name]",
		Run: func(cmd *cobra.Command, args []string) {
			handler.DeleteHandler(args)
		},
		Short: "Drop a service from the list of services that are being watched by sminit",
		Args:  cobra.ExactArgs(1),
	}

	var stopCmd = &cobra.Command{
		Use: "stop [service_name]",
		Run: func(cmd *cobra.Command, args []string) {
			handler.StopHandler(args)
		},
		Short: "Stop a running service",
		Args:  cobra.ExactArgs(1),
	}

	var logCmd = &cobra.Command{
		Use: "log",
		Run: func(cmd *cobra.Command, args []string) {
			handler.LogHandler()
		},
		Short: "Show sminit logs",
		Args:  cobra.ExactArgs(0),
	}

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(logCmd)
	_ = rootCmd.Execute()
}
