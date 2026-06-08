package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	_ "sigs.k8s.io/controller-runtime/pkg/client"
	_ "github.com/charmbracelet/bubbletea"
	_ "github.com/openshift/api/config/v1"
)

var rootCmd = &cobra.Command{
	Use:   "obstool",
	Short: "OpenShift Observability Tool",
	Long:  "A CLI tool for deploying and managing observability components on OpenShift",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("obstool - OpenShift Observability Tool")
		fmt.Println("Use --help for more information")
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
