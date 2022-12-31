package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "root [sub]",
		Short: "My root command",
	}

	// var subCmd = &cobra.Command{
	// 	Use:   "sub [no options!]",
	// 	Short: "My subcommand",
	// 	PreRun: func(cmd *cobra.Command, args []string) {
	// 		fmt.Printf("Inside subCmd PreRun with args: %v\n", args)
	// 	},
	// 	Run: func(cmd *cobra.Command, args []string) {
	// 		fmt.Printf("Inside subCmd Run with args: %v\n", args)
	// 	},
	// 	PostRun: func(cmd *cobra.Command, args []string) {
	// 		fmt.Printf("Inside subCmd PostRun with args: %v\n", args)
	// 	},
	// 	PersistentPostRun: func(cmd *cobra.Command, args []string) {
	// 		fmt.Printf("Inside subCmd PersistentPostRun with args: %v\n", args)
	// 	},
	// }

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "list: lists all subcommands",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("this is the list command\n")
		},
	}
	rootCmd.SetHelpTemplate("help template\n")
	rootCmd.AddCommand(listCmd)
	// rootCmd.SetArgs([]string{""})
	rootCmd.Execute()
	// fmt.Println()
	// rootCmd.SetArgs([]string{"sub", "arg1", "arg2"})
	// rootCmd.Execute()
	// c := exec.Command("ping", "google.com")
	// c.ProcessState.
}
