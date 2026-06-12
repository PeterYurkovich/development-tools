package cleanup

import (
	"github.com/spf13/cobra"
)

var CleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Cleanup and scale down observability components",
	Long:  "Commands for cleaning up and scaling down observability components",
}
