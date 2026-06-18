package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/observability-ui/development-tools/cmd/cleanup"
	"github.com/observability-ui/development-tools/cmd/deploy"
	"github.com/observability-ui/development-tools/cmd/update"
	"github.com/observability-ui/development-tools/cmd/users"
	execctx "github.com/observability-ui/development-tools/pkg/context"
	"github.com/observability-ui/development-tools/pkg/k8s"
)

var rootCmd = &cobra.Command{
	Use:   "obstool",
	Short: "OpenShift Observability Tool",
	Long:  "A unified CLI tool for deploying and managing observability components on OpenShift clusters",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().String("kubeconfig", "", "Path to kubeconfig file (defaults to $KUBECONFIG or ~/.kube/config)")

	rootCmd.AddCommand(update.UpdateCmd)
	rootCmd.AddCommand(cleanup.CleanupCmd)
	rootCmd.AddCommand(deploy.DeployCmd)
	rootCmd.AddCommand(users.UsersCmd)

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "version" {
			return nil
		}

		kubeconfigPath, _ := cmd.Flags().GetString("kubeconfig")

		kubeClient, err := k8s.NewClient(cmd.Context(), kubeconfigPath)
		if err != nil {
			return fmt.Errorf("failed to create kubernetes client: %w", err)
		}

		ctx := execctx.WithClient(cmd.Context(), kubeClient.Client)
		cmd.SetContext(ctx)

		return nil
	}
}
